# Copyright 2019 The Kubernetes Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

<#
.SYNOPSIS
  Library for configuring Windows nodes and joining them to the cluster.

.NOTES
  This module depends on common.psm1.

  Some portions copied / adapted from
  https://github.com/Microsoft/SDN/blob/master/Kubernetes/windows/start-kubelet.ps1.

.EXAMPLE
  Suggested usage for dev/test:
    [Net.ServicePointManager]::SecurityProtocol = `
        [Net.SecurityProtocolType]::Tls12
    Invoke-WebRequest `
        https://github.com/kubernetes/kubernetes/raw/master/cluster/gce/windows/k8s-node-setup.psm1 `
        -OutFile C:\k8s-node-setup.psm1
    Invoke-WebRequest `
        https://github.com/kubernetes/kubernetes/raw/master/cluster/gce/windows/configure.ps1 `
        -OutFile C:\configure.ps1
    Import-Module -Force C:\k8s-node-setup.psm1  # -Force to override existing
    # Execute functions manually or run configure.ps1.
#>

# TODO: update scripts for these style guidelines:
#  - Remove {} around variable references unless actually needed for clarity.
#  - Always use single-quoted strings unless actually interpolating variables
#    or using escape characters.
#  - Use "approved verbs":
#    https://docs.microsoft.com/en-us/powershell/developer/cmdlet/approved-verbs-for-windows-powershell-commands
#  - Document functions using proper syntax:
#    https://technet.microsoft.com/en-us/library/hh847834(v=wps.620).aspx

$INFRA_CONTAINER = "kubeletwin/pause"
$GCE_METADATA_SERVER = "169.254.169.254"
# The "management" interface is used by the kubelet and by Windows pods to talk
# to the rest of the Kubernetes cluster *without NAT*. This interface does not
# exist until an initial HNS network has been created on the Windows node - see
# Add_InitialHnsNetwork().
$MGMT_ADAPTER_NAME = "vEthernet (Ethernet*"

Import-Module -Force C:\common.psm1

# Writes a TODO with $Message to the console.
function Log_Todo {
  param (
    [parameter(Mandatory=$true)] [string]$Message
  )
  Log-Output "TODO: ${Message}"
}

# Writes a not-implemented warning with $Message to the console and exits the
# script.
function Log_NotImplemented {
  param (
    [parameter(Mandatory=$true)] [string]$Message
  )
  Log-Output "Not implemented yet: ${Message}" -Fatal
}

# Fails and exits if the route to the GCE metadata server is not present,
# otherwise does nothing and emits nothing.
function Verify_GceMetadataServerRouteIsPresent {
  Try {
    Get-NetRoute `
        -ErrorAction "Stop" `
        -AddressFamily IPv4 `
        -DestinationPrefix ${GCE_METADATA_SERVER}/32 | Out-Null
  } Catch [Microsoft.PowerShell.Cmdletization.Cim.CimJobException] {
    Log-Output -Fatal `
        ("GCE metadata server route is not present as expected.`n" +
         "$(Get-NetRoute -AddressFamily IPv4 | Out-String)")
  }
}

# Checks if the route to the GCE metadata server is present. Returns when the
# route is NOT present or after a timeout has expired.
function WaitFor_GceMetadataServerRouteToBeRemoved {
  $elapsed = 0
  $timeout = 60
  Log-Output ("Waiting up to ${timeout} seconds for GCE metadata server " +
              "route to be removed")
  while (${elapsed} -lt ${timeout}) {
    Try {
      Get-NetRoute `
          -ErrorAction "Stop" `
          -AddressFamily IPv4 `
          -DestinationPrefix ${GCE_METADATA_SERVER}/32 | Out-Null
    } Catch [Microsoft.PowerShell.Cmdletization.Cim.CimJobException] {
      break
    }
    $sleeptime = 2
    Start-Sleep ${sleeptime}
    ${elapsed} += ${sleeptime}
  }
}

# Adds a route to the GCE metadata server to every network interface.
function Add_GceMetadataServerRoute {
  # Before setting up HNS the Windows VM has a "vEthernet (nat)" interface and
  # a "Ethernet" interface, and the route to the metadata server exists on the
  # Ethernet interface. After adding the HNS network a "vEthernet (Ethernet)"
  # interface is added, and it seems to subsume the routes of the "Ethernet"
  # interface (trying to add routes on the Ethernet interface at this point just
  # results in "New-NetRoute : Element not found" errors). I don't know what's
  # up with that, but since it's hard to know what's the right thing to do here
  # we just try to add the route on all of the network adapters.
  Get-NetAdapter | ForEach-Object {
    $adapter_index = $_.InterfaceIndex
    New-NetRoute `
        -ErrorAction Ignore `
        -DestinationPrefix "${GCE_METADATA_SERVER}/32" `
        -InterfaceIndex ${adapter_index} | Out-Null
  }
}

# Fetches the kube-env from the instance metadata.
#
# Returns: a PowerShell Hashtable object containing the key-value pairs from
#   kube-env.
function Fetch-KubeEnv {
  # Testing / debugging:
  # First:
  #   ${kube_env} = Get-InstanceMetadataValue 'kube-env'
  # or:
  #   ${kube_env} = [IO.File]::ReadAllText(".\kubeEnv.txt")
  # ${kube_env_table} = ConvertFrom-Yaml ${kube_env}
  # ${kube_env_table}
  # ${kube_env_table}.GetType()

  # The type of kube_env is a powershell String.
  $kube_env = Get-InstanceMetadataValue 'kube-env'
  $kube_env_table = ConvertFrom-Yaml ${kube_env}
  return ${kube_env_table}
}

# Sets the environment variable $Key to $Value at the Machine scope (will
# be present in the environment for all new shells after a reboot).
function Set_MachineEnvironmentVar {
  param (
    [parameter(Mandatory=$true)] [string]$Key,
    [parameter(Mandatory=$true)] [string]$Value
  )
  [Environment]::SetEnvironmentVariable($Key, $Value, "Machine")
}

# Sets the environment variable $Key to $Value in the current shell.
function Set_CurrentShellEnvironmentVar {
  param (
    [parameter(Mandatory=$true)] [string]$Key,
    [parameter(Mandatory=$true)] [string]$Value
  )
  $expression = '$env:' + $Key + ' = "' + $Value + '"'
  Invoke-Expression ${expression}
}

# Sets environment variables used by Kubernetes binaries and by other functions
# in this module. Depends on numerous ${kube_env} keys.
function Set-EnvironmentVars {
  # Turning the kube-env values into environment variables is not required but
  # it makes debugging this script easier, and it also makes the syntax a lot
  # easier (${env:K8S_DIR} can be expanded within a string but
  # ${kube_env}['K8S_DIR'] cannot be afaik).
  $env_vars = @{
    "K8S_DIR" = ${kube_env}['K8S_DIR']
    "NODE_DIR" = ${kube_env}['NODE_DIR']
    "CNI_DIR" = ${kube_env}['CNI_DIR']
    "CNI_CONFIG_DIR" = ${kube_env}['CNI_CONFIG_DIR']
    "PKI_DIR" = ${kube_env}['PKI_DIR']
    "KUBELET_CONFIG" = ${kube_env}['KUBELET_CONFIG_FILE']
    "BOOTSTRAP_KUBECONFIG" = ${kube_env}['BOOTSTRAP_KUBECONFIG_FILE']
    "KUBEPROXY_KUBECONFIG" = ${kube_env}['KUBEPROXY_KUBECONFIG_FILE']

    "Path" = ${env:Path} + ";" + ${kube_env}['NODE_DIR']
    "KUBE_NETWORK" = "l2bridge".ToLower()
    "CA_CERT_BUNDLE_PATH" = ${kube_env}['PKI_DIR'] + '\ca-certificates.crt'
    "KUBELET_CERT_PATH" = ${kube_env}['PKI_DIR'] + '\kubelet.crt'
    "KUBELET_KEY_PATH" = ${kube_env}['PKI_DIR'] + '\kubelet.key'

    # TODO(pjh): these are only in flags, can be removed from env once flags are
    # moved to util.sh:
    "LOGS_DIR" = ${kube_env}['LOGS_DIR']
    "MANIFESTS_DIR" = ${kube_env}['MANIFESTS_DIR']
    "KUBECONFIG" = ${kube_env}['KUBECONFIG_FILE']
  }

  # Set the environment variables in two ways: permanently on the machine (only
  # takes effect after a reboot), and in the current shell.
  $env_vars.GetEnumerator() | ForEach-Object{
    $message = "Setting environment variable: " + $_.key + " = " + $_.value
    Log-Output ${message}
    Set_MachineEnvironmentVar $_.key $_.value
    Set_CurrentShellEnvironmentVar $_.key $_.value
  }
}

# Configures various settings and prerequisites needed for the rest of the
# functions in this module and the Kubernetes binaries to operate properly.
function Set-PrerequisiteOptions {
  # The Windows firewall interferes with Kubernetes networking; GCE's firewall
  # should be sufficient.
  Log-Output "Disabling Windows Firewall"
  Set-NetFirewallProfile -Profile Domain, Public, Private -Enabled False

  # Windows updates cause the node to reboot at arbitrary times.
  Log-Output "Disabling Windows Update service"
  sc.exe config wuauserv start=disabled
  sc.exe stop wuauserv

  # Windows Defender periodically consumes 100% of the CPU.
  # TODO(pjh): this (all of a sudden, ugh) started failing with "The term
  # 'Set-MpPreference' is not recognized...". Investigate and fix or remove.
  #Log-Output "Disabling Windows Defender service"
  #Set-MpPreference -DisableRealtimeMonitoring $true
  #Uninstall-WindowsFeature -Name 'Windows-Defender'

  # Use TLS 1.2: needed for Invoke-WebRequest downloads from github.com.
  [Net.ServicePointManager]::SecurityProtocol = `
      [Net.SecurityProtocolType]::Tls12

  # https://github.com/cloudbase/powershell-yaml
  Log-Output "Installing powershell-yaml module from external repo"
  Install-Module -Name powershell-yaml -Force
}

# Creates directories where other functions in this module will read and write
# data.
function Create-Directories {
  Log-Output "Creating ${env:K8S_DIR} and its subdirectories."
  ForEach ($dir in ("${env:K8S_DIR}", "${env:NODE_DIR}", "${env:LOGS_DIR}",
    "${env:CNI_DIR}", "${env:CNI_CONFIG_DIR}", "${env:MANIFESTS_DIR}",
    "${env:PKI_DIR}")) {
    mkdir -Force $dir
  }
}

# Downloads some external helper scripts needed by other functions in this
# module.
function Download-HelperScripts {
  if (-not (ShouldWrite-File ${env:K8S_DIR}\hns.psm1)) {
    return
  }
  MustDownload-File -OutFile ${env:K8S_DIR}\hns.psm1 `
    -URLs "https://github.com/Microsoft/SDN/raw/master/Kubernetes/windows/hns.psm1"
}

# Takes the Windows version string from the cluster bash scripts (e.g.
# 'win1809') and returns the correct label to use for containers on this
# version of Windows. Returns $null if $WinVersion is unknown.
function Get_ContainerVersionLabel {
  param (
    [parameter(Mandatory=$true)] [string]$WinVersion
  )
  # -match does regular expression matching.
  if ($WinVersion -match '1809') {
    return '1809'
  }
  elseif ($WinVersion -match '2019') {
    return 'ltsc2019'
  }
  Throw ("Unknown Windows version $WinVersion, don't know its container " +
         "version label")
}

# Builds the pause image with name $INFRA_CONTAINER.
function Create-PauseImage {
  $win_version = $(Get-InstanceMetadataValue 'win-version')
  if ($win_version -match '2019') {
    # TODO(pjh): update this function to properly support 2019 vs. 1809 vs.
    # future Windows versions. For example, Windows Server 2019 does not
    # support the nanoserver container
    # (https://blogs.technet.microsoft.com/virtualization/2018/11/13/windows-server-2019-now-available/).
    Log_NotImplemented "Need to update Create-PauseImage for WS2019"
  }

  $version_label = Get_ContainerVersionLabel $win_version
  $pause_dir = "${env:K8S_DIR}\pauseimage"
  $dockerfile = "$pause_dir\Dockerfile"
  mkdir -Force $pause_dir
  if (ShouldWrite-File $dockerfile) {
    New-Item -Force -ItemType file $dockerfile | Out-Null
    Set-Content `
        $dockerfile `
        ("FROM mcr.microsoft.com/windows/nanoserver:${version_label}`n`n" +
         "CMD cmd /c ping -t localhost > nul")
  }

  if (($(docker images -a) -like "*${INFRA_CONTAINER}*") -and
      (-not $REDO_STEPS)) {
    Log-Output "Skip: ${INFRA_CONTAINER} already built"
    return
  }
  docker build -t ${INFRA_CONTAINER} $pause_dir
  if ($LastExitCode -ne 0) {
    Log-Output -Fatal `
        "docker build -t ${INFRA_CONTAINER} $pause_dir failed"
  }
}

# Downloads the Kubernetes binaries from kube-env's NODE_BINARY_TAR_URL and
# puts them in a subdirectory of $env:K8S_DIR.
#
# Required ${kube_env} keys:
#   NODE_BINARY_TAR_URL
function DownloadAndInstall-KubernetesBinaries {
  # Assume that presence of kubelet.exe indicates that the kubernetes binaries
  # were already previously downloaded to this node.
  if (-not (ShouldWrite-File ${env:NODE_DIR}\kubelet.exe)) {
    return
  }

  $tmp_dir = 'C:\k8s_tmp'
  New-Item -Force -ItemType 'directory' $tmp_dir | Out-Null

  $urls = ${kube_env}['NODE_BINARY_TAR_URL'].Split(",")
  $filename = Split-Path -leaf $urls[0]
  $hash = $null
  if ($kube_env.ContainsKey('NODE_BINARY_TAR_HASH')) {
    $hash = ${kube_env}['NODE_BINARY_TAR_HASH']
  }
  MustDownload-File -Hash $hash -OutFile ${tmp_dir}\${filename} -URLs $urls

  # Change the directory to the parent directory of ${env:K8S_DIR} and untar.
  # This (over-)writes ${dest_dir}/kubernetes/node/bin/*.exe files.
  $dest_dir = (Get-Item ${env:K8S_DIR}).Parent.Fullname
  tar xzf ${tmp_dir}\${filename} -C ${dest_dir}

  # Clean up the temporary directory
  Remove-Item -Force -Recurse $tmp_dir
}

# TODO(pjh): this is copied from
# https://github.com/Microsoft/SDN/blob/master/Kubernetes/windows/start-kubelet.ps1#L98.
# See if there's a way to fetch or construct the "management subnet" so that
# this is not needed.
function ConvertTo_DecimalIP
{
  param(
    [parameter(Mandatory = $true, Position = 0)]
    [Net.IPAddress] $IPAddress
  )

  $i = 3; $decimal_ip = 0;
  $IPAddress.GetAddressBytes() | % {
    $decimal_ip += $_ * [Math]::Pow(256, $i); $i--
  }
  return [UInt32]$decimal_ip
}

# TODO(pjh): this is copied from
# https://github.com/Microsoft/SDN/blob/master/Kubernetes/windows/start-kubelet.ps1#L98.
# See if there's a way to fetch or construct the "management subnet" so that
# this is not needed.
function ConvertTo_DottedDecimalIP
{
  param(
    [parameter(Mandatory = $true, Position = 0)]
    [Uint32] $IPAddress
  )

  $dotted_ip = $(for ($i = 3; $i -gt -1; $i--) {
    $remainder = $IPAddress % [Math]::Pow(256, $i)
    ($IPAddress - $remainder) / [Math]::Pow(256, $i)
    $IPAddress = $remainder
  })
  return [String]::Join(".", $dotted_ip)
}

# TODO(pjh): this is copied from
# https://github.com/Microsoft/SDN/blob/master/Kubernetes/windows/start-kubelet.ps1#L98.
# See if there's a way to fetch or construct the "management subnet" so that
# this is not needed.
function ConvertTo_MaskLength
{
  param(
    [parameter(Mandatory = $True, Position = 0)]
    [Net.IPAddress] $SubnetMask
  )

  $bits = "$($SubnetMask.GetAddressBytes() | % {
    [Convert]::ToString($_, 2)
  } )" -replace "[\s0]"
  return $bits.Length
}

# Returns the "management" subnet on which the Windows pods+kubelet will
# communicate with the rest of the Kubernetes cluster without NAT. In GCE this
# is the subnet that VM internal IPs are allocated from.
#
# This function will fail if Add_InitialHnsNetwork() has not been called first.
function Get_MgmtSubnet {
  $net_adapter = Get_MgmtNetAdapter

  # TODO(pjh): applying the primary interface's subnet mask to its IP address
  # *should* give us the GCE network subnet that VM IP addresses are being
  # allocated from... however it might be more accurate or straightforward to
  # just fetch the IP address range for the VPC subnet that the kube-up script
  # creates (kubernetes-subnet-default).
  $addr = (Get-NetIPAddress `
      -InterfaceAlias ${net_adapter}.ifAlias `
      -AddressFamily IPv4).IPAddress
  $mask = (Get-WmiObject Win32_NetworkAdapterConfiguration |
      Where-Object InterfaceIndex -eq $(${net_adapter}.ifIndex)).IPSubnet[0]
  $mgmt_subnet = `
    (ConvertTo_DecimalIP ${addr}) -band (ConvertTo_DecimalIP ${mask})
  $mgmt_subnet = ConvertTo_DottedDecimalIP ${mgmt_subnet}
  return "${mgmt_subnet}/$(ConvertTo_MaskLength $mask)"
}

# Returns a network adapter object for the "management" interface via which the
# Windows pods+kubelet will communicate with the rest of the Kubernetes cluster.
#
# This function will fail if Add_InitialHnsNetwork() has not been called first.
function Get_MgmtNetAdapter {
  $net_adapter = Get-NetAdapter | Where-Object Name -like ${MGMT_ADAPTER_NAME}
  if (-not ${net_adapter}) {
    Throw ("Failed to find a suitable network adapter, check your network " +
           "settings.")
  }

  return $net_adapter
}

# Decodes the base64 $Data string and writes it as binary to $File. Does
# nothing if $File already exists and $REDO_STEPS is not set.
function Write_PkiData {
  param (
    [parameter(Mandatory=$true)] [string] $Data,
    [parameter(Mandatory=$true)] [string] $File
  )

  if (-not (ShouldWrite-File $File)) {
    return
  }

  # This command writes out a PEM certificate file, analogous to "base64
  # --decode" on Linux. See https://stackoverflow.com/a/51914136/1230197.
  [IO.File]::WriteAllBytes($File, [Convert]::FromBase64String($Data))
  Log_Todo ("need to set permissions correctly on ${File}; not sure what the " +
            "Windows equivalent of 'umask 077' is")
  # Linux: owned by root, rw by user only.
  #   -rw------- 1 root root 1.2K Oct 12 00:56 ca-certificates.crt
  #   -rw------- 1 root root 1.3K Oct 12 00:56 kubelet.crt
  #   -rw------- 1 root root 1.7K Oct 12 00:56 kubelet.key
  # Windows:
  #   https://docs.microsoft.com/en-us/dotnet/api/system.io.fileattributes
  #   https://docs.microsoft.com/en-us/dotnet/api/system.io.fileattributes
}

# Creates the node PKI files in $env:PKI_DIR.
#
# Required ${kube_env} keys:
#   CA_CERT
#   KUBELET_CERT
#   KUBELET_KEY
function Create-NodePki {
  Log-Output "Creating node pki files"

  $CA_CERT_BUNDLE = ${kube_env}['CA_CERT']
  $KUBELET_CERT = ${kube_env}['KUBELET_CERT']
  $KUBELET_KEY = ${kube_env}['KUBELET_KEY']

  Write_PkiData "${CA_CERT_BUNDLE}" ${env:CA_CERT_BUNDLE_PATH}
  Write_PkiData "${KUBELET_CERT}" ${env:KUBELET_CERT_PATH}
  Write_PkiData "${KUBELET_KEY}" ${env:KUBELET_KEY_PATH}
  Get-ChildItem ${env:PKI_DIR}
}

# Creates the kubelet kubeconfig at $env:BOOTSTRAP_KUBECONFIG.
#
# Create-NodePki() must be called first.
#
# Required ${kube_env} keys:
#   KUBERNETES_MASTER_NAME: the apiserver IP address.
function Create-KubeletKubeconfig {
  # The API server IP address comes from KUBERNETES_MASTER_NAME in kube-env, I
  # think. cluster/gce/gci/configure-helper.sh?l=2801
  $apiserverAddress = ${kube_env}['KUBERNETES_MASTER_NAME']

  # TODO(pjh): set these using kube-env values.
  $createBootstrapConfig = $true
  $fetchBootstrapConfig = $false

  if (${createBootstrapConfig}) {
    if (-not (ShouldWrite-File ${env:BOOTSTRAP_KUBECONFIG})) {
      return
    }
    New-Item -Force -ItemType file ${env:BOOTSTRAP_KUBECONFIG} | Out-Null
    # TODO(mtaufen): is user "kubelet" correct? Other examples use e.g.
    #   "system:node:$(hostname)".
    Set-Content ${env:BOOTSTRAP_KUBECONFIG} `
'apiVersion: v1
kind: Config
users:
- name: kubelet
  user:
    client-certificate: KUBELET_CERT_PATH
    client-key: KUBELET_KEY_PATH
clusters:
- name: local
  cluster:
    server: https://APISERVER_ADDRESS
    certificate-authority: CA_CERT_BUNDLE_PATH
contexts:
- context:
    cluster: local
    user: kubelet
  name: service-account-context
current-context: service-account-context'.`
      replace('KUBELET_CERT_PATH', ${env:KUBELET_CERT_PATH}).`
      replace('KUBELET_KEY_PATH', ${env:KUBELET_KEY_PATH}).`
      replace('APISERVER_ADDRESS', ${apiserverAddress}).`
      replace('CA_CERT_BUNDLE_PATH', ${env:CA_CERT_BUNDLE_PATH})
    Log-Output ("kubelet bootstrap kubeconfig:`n" +
                "$(Get-Content -Raw ${env:BOOTSTRAP_KUBECONFIG})")
  }
  elseif (${fetchBootstrapConfig}) {
    Log_NotImplemented `
        "fetching kubelet bootstrap-kubeconfig file from metadata"
    # get-metadata-value "instance/attributes/bootstrap-kubeconfig" >
    #   /var/lib/kubelet/bootstrap-kubeconfig
    Log-Output ("kubelet bootstrap kubeconfig:`n" +
                "$(Get-Content -Raw ${env:BOOTSTRAP_KUBECONFIG})")
  }
  else {
    Log_NotImplemented "fetching kubelet kubeconfig file from metadata"
  }
}

# Creates the kube-proxy user kubeconfig file at $env:KUBEPROXY_KUBECONFIG.
#
# Create-NodePki() must be called first.
#
# Required ${kube_env} keys:
#   CA_CERT
#   KUBE_PROXY_TOKEN
function Create-KubeproxyKubeconfig {
  if (-not (ShouldWrite-File ${env:KUBEPROXY_KUBECONFIG})) {
    return
  }

  New-Item -Force -ItemType file ${env:KUBEPROXY_KUBECONFIG} | Out-Null

  # In configure-helper.sh kubelet kubeconfig uses certificate-authority while
  # kubeproxy kubeconfig uses certificate-authority-data, ugh. Does it matter?
  # Use just one or the other for consistency?
  Set-Content ${env:KUBEPROXY_KUBECONFIG} `
'apiVersion: v1
kind: Config
users:
- name: kube-proxy
  user:
    token: KUBEPROXY_TOKEN
clusters:
- name: local
  cluster:
    certificate-authority-data: CA_CERT
contexts:
- context:
    cluster: local
    user: kube-proxy
  name: service-account-context
current-context: service-account-context'.`
    replace('KUBEPROXY_TOKEN', ${kube_env}['KUBE_PROXY_TOKEN']).`
    replace('CA_CERT', ${kube_env}['CA_CERT'])

  Log-Output ("kubeproxy kubeconfig:`n" +
              "$(Get-Content -Raw ${env:KUBEPROXY_KUBECONFIG})")
}

# Returns the IP alias range configured for this GCE instance.
function Get_IpAliasRange {
  $url = ("http://${GCE_METADATA_SERVER}/computeMetadata/v1/instance/" +
          "network-interfaces/0/ip-aliases/0")
  $client = New-Object Net.WebClient
  $client.Headers.Add('Metadata-Flavor', 'Google')
  return ($client.DownloadString($url)).Trim()
}

# Retrieves the pod CIDR and sets it in $env:POD_CIDR.
function Set-PodCidr {
  while($true) {
    $pod_cidr = Get_IpAliasRange
    if (-not $?) {
      Log-Output ${pod_cIDR}
      Log-Output "Retrying Get_IpAliasRange..."
      Start-Sleep -sec 1
      continue
    }
    break
  }

  Log-Output "fetched pod CIDR (same as IP alias range): ${pod_cidr}"
  Set_MachineEnvironmentVar "POD_CIDR" ${pod_cidr}
  Set_CurrentShellEnvironmentVar "POD_CIDR" ${pod_cidr}
}

# Adds an initial HNS network on the Windows node which forces the creation of
# a virtual switch and the "management" interface that will be used to
# communicate with the rest of the Kubernetes cluster without NAT.
#
# Note that adding the initial HNS network may cause connectivity to the GCE
# metadata server to be lost due to a Windows bug.
# Configure-HostNetworkingService() restores connectivity, look there for
# details.
#
# Download-HelperScripts() must have been called first.
function Add_InitialHnsNetwork {
  $INITIAL_HNS_NETWORK = 'External'

  # This comes from
  # https://github.com/Microsoft/SDN/blob/master/Kubernetes/flannel/l2bridge/start.ps1#L74
  # (or
  # https://github.com/Microsoft/SDN/blob/master/Kubernetes/windows/start-kubelet.ps1#L206).
  #
  # daschott noted on Slack: "L2bridge networks require an external vSwitch.
  # The first network ("External") with hardcoded values in the script is just
  # a placeholder to create an external vSwitch. This is purely for convenience
  # to be able to remove/modify the actual HNS network ("cbr0") or rejoin the
  # nodes without a network blip. Creating a vSwitch takes time, causes network
  # blips, and it makes it more likely to hit the issue where flanneld is
  # stuck, so we want to do this as rarely as possible."
  $hns_network = Get-HnsNetwork | Where-Object Name -eq $INITIAL_HNS_NETWORK
  if ($hns_network) {
    if ($REDO_STEPS) {
      Log-Output ("Warning: initial '$INITIAL_HNS_NETWORK' HNS network " +
                  "already exists, removing it and recreating it")
      $hns_network | Remove-HnsNetwork
      $hns_network = $null
    }
    else {
      Log-Output ("Skip: initial '$INITIAL_HNS_NETWORK' HNS network " +
                  "already exists, not recreating it")
      return
    }
  }
  Log-Output ("Creating initial HNS network to force creation of " +
              "${MGMT_ADAPTER_NAME} interface")
  # Note: RDP connection will hiccup when running this command.
  New-HNSNetwork `
      -Type "L2Bridge" `
      -AddressPrefix "192.168.255.0/30" `
      -Gateway "192.168.255.1" `
      -Name $INITIAL_HNS_NETWORK `
      -Verbose
}

# Configures HNS on the Windows node to enable Kubernetes networking:
#   - Creates the "management" interface associated with an initial HNS network.
#   - Creates the HNS network $env:KUBE_NETWORK for pod networking.
#   - Creates an HNS endpoint for pod networking.
#   - Adds necessary routes on the management interface.
#   - Verifies that the GCE metadata server connection remains intact.
#
# Prerequisites:
#   $env:POD_CIDR is set (by Set-PodCidr).
#   Download-HelperScripts() has been called.
function Configure-HostNetworkingService {
  Import-Module -Force ${env:K8S_DIR}\hns.psm1

  Add_InitialHnsNetwork

  # For Windows nodes the pod gateway IP address is the .1 address in the pod
  # CIDR for the host, but from inside containers it's the .2 address.
  $pod_gateway = `
      ${env:POD_CIDR}.substring(0, ${env:POD_CIDR}.lastIndexOf('.')) + '.1'
  $pod_endpoint_gateway = `
      ${env:POD_CIDR}.substring(0, ${env:POD_CIDR}.lastIndexOf('.')) + '.2'
  Log-Output ("Setting up Windows node HNS networking: " +
              "podCidr = ${env:POD_CIDR}, podGateway = ${pod_gateway}, " +
              "podEndpointGateway = ${pod_endpoint_gateway}")

  $hns_network = Get-HnsNetwork | Where-Object Name -eq ${env:KUBE_NETWORK}
  if ($hns_network) {
    if ($REDO_STEPS) {
      Log-Output ("Warning: ${env:KUBE_NETWORK} HNS network already exists, " +
                  "removing it and recreating it")
      $hns_network | Remove-HnsNetwork
      $hns_network = $null
    }
    else {
      Log-Output "Skip: ${env:KUBE_NETWORK} HNS network already exists"
    }
  }
  $created_hns_network = $false
  if (-not $hns_network) {
    # Note: RDP connection will hiccup when running this command.
    $hns_network = New-HNSNetwork `
        -Type "L2Bridge" `
        -AddressPrefix ${env:POD_CIDR} `
        -Gateway ${pod_gateway} `
        -Name ${env:KUBE_NETWORK} `
        -Verbose
    $created_hns_network = $true
  }

  $endpoint_name = "cbr0"
  $vnic_name = "vEthernet (${endpoint_name})"

  $hns_endpoint = Get-HnsEndpoint | Where-Object Name -eq $endpoint_name
  # Note: we don't expect to ever enter this block currently - while the HNS
  # network does seem to persist across reboots, the HNS endpoints do not.
  if ($hns_endpoint) {
    if ($REDO_STEPS) {
      Log-Output ("Warning: HNS endpoint $endpoint_name already exists, " +
                  "removing it and recreating it")
      $hns_endpoint | Remove-HnsEndpoint
      $hns_endpoint = $null
    }
    else {
      Log-Output "Skip: HNS endpoint $endpoint_name already exists"
    }
  }
  if (-not $hns_endpoint) {
    $hns_endpoint = New-HnsEndpoint `
        -NetworkId ${hns_network}.Id `
        -Name ${endpoint_name} `
        -IPAddress ${pod_endpoint_gateway} `
        -Gateway "0.0.0.0" `
        -Verbose
    # TODO(pjh): find out: why is this always CompartmentId 1?
    Attach-HnsHostEndpoint `
        -EndpointID ${hns_endpoint}.Id `
        -CompartmentID 1 `
        -Verbose
    netsh interface ipv4 set interface "${vnic_name}" forwarding=enabled
  }

  Get-HNSPolicyList | Remove-HnsPolicyList

  # Add a route from the management NIC to the pod CIDR.
  #
  # When a packet from a Kubernetes service backend arrives on the destination
  # Windows node, the reverse SNAT will be applied and the source address of
  # the packet gets replaced from the pod IP to the service VIP. The packet
  # will then leave the VM and return back through hairpinning.
  #
  # When IP alias is enabled, IP forwarding is disabled for anti-spoofing;
  # the packet with the service VIP will get blocked and be lost. With this
  # route, the packet will be routed to the pod subnetwork, and not leave the
  # VM.
  $mgmt_net_adapter = Get_MgmtNetAdapter
  New-NetRoute `
      -ErrorAction Ignore `
      -InterfaceAlias ${mgmt_net_adapter}.ifAlias `
      -DestinationPrefix ${env:POD_CIDR} `
      -NextHop "0.0.0.0" `
      -Verbose

  if ($created_hns_network) {
    # There is an HNS bug where the route to the GCE metadata server will be
    # removed when the HNS network is created:
    # https://github.com/Microsoft/hcsshim/issues/299#issuecomment-425491610.
    # The behavior here is very unpredictable: the route may only be removed
    # after some delay, or it may appear to be removed then you'll add it back
    # but then it will be removed once again. So, we first wait a long
    # unfortunate amount of time to ensure that things have quiesced, then we
    # wait until we're sure the route is really gone before re-adding it again.
    Log-Output "Waiting 45 seconds for host network state to quiesce"
    Start-Sleep 45
    WaitFor_GceMetadataServerRouteToBeRemoved
    Log-Output "Re-adding the GCE metadata server route"
    Add_GceMetadataServerRoute
  }
  Verify_GceMetadataServerRouteIsPresent

  Log-Output "Host network setup complete"
}

# Downloads the Windows CNI binaries and writes a CNI config file under
# $env:CNI_CONFIG_DIR.
#
# Prerequisites:
#   $env:POD_CIDR is set (by Set-PodCidr).
#   The "management" interface exists (Configure-HostNetworkingService).
#   The HNS network for pod networking has been configured
#     (Configure-HostNetworkingService).
#
# Required ${kube_env} keys:
#   DNS_SERVER_IP
#   DNS_DOMAIN
#   CLUSTER_IP_RANGE
#   SERVICE_CLUSTER_IP_RANGE
function Configure-CniNetworking {
  if ((ShouldWrite-File ${env:CNI_DIR}\win-bridge.exe) -or
      (ShouldWrite-File ${env:CNI_DIR}\host-local.exe)) {
    MustDownload-File -OutFile ${env:CNI_DIR}\windows-cni-plugins.zip `
      -URLs "https://github.com/yujuhong/gce-k8s-windows-testing/raw/master/windows-cni-plugins.zip"
    rm ${env:CNI_DIR}\*.exe
    Expand-Archive ${env:CNI_DIR}\windows-cni-plugins.zip ${env:CNI_DIR}
    mv ${env:CNI_DIR}\bin\*.exe ${env:CNI_DIR}\
    rmdir ${env:CNI_DIR}\bin
  }
  if (-not ((Test-Path ${env:CNI_DIR}\win-bridge.exe) -and `
            (Test-Path ${env:CNI_DIR}\host-local.exe))) {
    Log-Output `
        "win-bridge.exe and host-local.exe not found in ${env:CNI_DIR}" `
        -Fatal
  }

  $l2bridge_conf = "${env:CNI_CONFIG_DIR}\l2bridge.conf"
  if (-not (ShouldWrite-File ${l2bridge_conf})) {
    return
  }

  $mgmt_ip = (Get_MgmtNetAdapter |
              Get-NetIPAddress -AddressFamily IPv4).IPAddress
  $mgmt_subnet = Get_MgmtSubnet
  Log-Output ("using mgmt IP ${mgmt_ip} and mgmt subnet ${mgmt_subnet} for " +
              "CNI config")

  # Explanation of the CNI config values:
  #   CLUSTER_CIDR: the cluster CIDR from which pod CIDRs are allocated.
  #   POD_CIDR: the pod CIDR assigned to this node.
  #   MGMT_SUBNET: the subnet on which the Windows pods + kubelet will
  #     communicate with the rest of the cluster without NAT (i.e. the subnet
  #     that VM internal IPs are allocated from).
  #   MGMT_IP: the IP address assigned to the node's primary network interface
  #     (i.e. the internal IP of the GCE VM).
  #   SERVICE_CIDR: the CIDR used for kubernetes services.
  #   DNS_SERVER_IP: the cluster's DNS server IP address.
  #   DNS_DOMAIN: the cluster's DNS domain, e.g. "cluster.local".
  New-Item -Force -ItemType file ${l2bridge_conf} | Out-Null
  Set-Content ${l2bridge_conf} `
'{
  "cniVersion":  "0.2.0",
  "name":  "l2bridge",
  "type":  "win-bridge",
  "capabilities":  {
    "portMappings":  true
  },
  "ipam":  {
    "type": "host-local",
    "subnet": "POD_CIDR"
  },
  "dns":  {
    "Nameservers":  [
      "DNS_SERVER_IP"
    ],
    "Search": [
      "DNS_DOMAIN"
    ]
  },
  "Policies":  [
    {
      "Name":  "EndpointPolicy",
      "Value":  {
        "Type":  "OutBoundNAT",
        "ExceptionList":  [
          "CLUSTER_CIDR",
          "SERVICE_CIDR",
          "MGMT_SUBNET"
        ]
      }
    },
    {
      "Name":  "EndpointPolicy",
      "Value":  {
        "Type":  "ROUTE",
        "DestinationPrefix":  "SERVICE_CIDR",
        "NeedEncap":  true
      }
    },
    {
      "Name":  "EndpointPolicy",
      "Value":  {
        "Type":  "ROUTE",
        "DestinationPrefix":  "MGMT_IP/32",
        "NeedEncap":  true
      }
    }
  ]
}'.replace('POD_CIDR', ${env:POD_CIDR}).`
  replace('DNS_SERVER_IP', ${kube_env}['DNS_SERVER_IP']).`
  replace('DNS_DOMAIN', ${kube_env}['DNS_DOMAIN']).`
  replace('MGMT_IP', ${mgmt_ip}).`
  replace('CLUSTER_CIDR', ${kube_env}['CLUSTER_IP_RANGE']).`
  replace('SERVICE_CIDR', ${kube_env}['SERVICE_CLUSTER_IP_RANGE']).`
  replace('MGMT_SUBNET', ${mgmt_subnet})

  Log-Output "CNI config:`n$(Get-Content -Raw ${l2bridge_conf})"
}

# Fetches the kubelet config from the instance metadata and puts it at
# $env:KUBELET_CONFIG.
function Configure-Kubelet {
  if (-not (ShouldWrite-File ${env:KUBELET_CONFIG})) {
    return
  }

  # The Kubelet config is built by build-kubelet-config() in
  # cluster/gce/util.sh, and stored in the metadata server under the
  # 'kubelet-config' key.
  $kubelet_config = Get-InstanceMetadataValue 'kubelet-config'
  Set-Content ${env:KUBELET_CONFIG} $kubelet_config
  Log-Output "Kubelet config:`n$(Get-Content -Raw ${env:KUBELET_CONFIG})"
}

# Sets up the kubelet and kube-proxy arguments and starts them as native
# Windows services.
#
# Required ${kube_env} keys:
#   KUBELET_ARGS
#   KUBERNETES_MASTER_NAME
#   CLUSTER_IP_RANGE
function Start-WorkerServices {
  $kubelet_args_str = ${kube_env}['KUBELET_ARGS']
  $kubelet_args = $kubelet_args_str.Split(" ")
  Log-Output "kubelet_args from metadata: ${kubelet_args}"

  $additional_arg_list = @(`
      "--pod-infra-container-image=${INFRA_CONTAINER}"
  )
  $kubelet_args = ${kubelet_args} + ${additional_arg_list}

  # kubeproxy is started on Linux nodes using
  # kube-manifests/kubernetes/gci-trusty/kube-proxy.manifest, which is
  # generated by start-kube-proxy in configure-helper.sh and contains e.g.:
  #   kube-proxy --master=https://35.239.84.171
  #   --kubeconfig=/var/lib/kube-proxy/kubeconfig --cluster-cidr=10.64.0.0/14
  #   --resource-container="" --oom-score-adj=-998 --v=2
  #   --feature-gates=ExperimentalCriticalPodAnnotation=true
  #   --iptables-sync-period=1m --iptables-min-sync-period=10s
  #   --ipvs-sync-period=1m --ipvs-min-sync-period=10s
  # And also with various volumeMounts and "securityContext: privileged: true".
  $apiserver_address = ${kube_env}['KUBERNETES_MASTER_NAME']
  $kubeproxy_args = @(`
      "--v=4",
      "--master=https://${apiserver_address}",
      "--kubeconfig=${env:KUBEPROXY_KUBECONFIG}",
      "--proxy-mode=kernelspace",
      "--hostname-override=$(hostname)",
      "--cluster-cidr=$(${kube_env}['CLUSTER_IP_RANGE'])",

      # Configure kube-proxy to run as a windows service.
      "--windows-service=true",

      # TODO(mtaufen): Configure logging for kube-proxy running as a service.
      # I haven't been able to figure out how to direct stdout/stderr into log
      # files when configuring it to run via sc.exe, so we just manually
      # override logging config here.
      "--log-file=${env:LOGS_DIR}\kube-proxy.log",
      # klog sets this to true intenrally, so need to override to false
      # so we actually log to the file
      "--logtostderr=false",

      # Configure flags with explicit empty string values. We can't escape
      # double-quotes, because they still break sc.exe after expansion in the
      # binPath parameter, and single-quotes get parsed as characters instead
      # of string delimiters.
      "--resource-container="
  )

  # TODO(pjh): kubelet is emitting these messages:
  # I1023 23:44:11.761915    2468 kubelet.go:274] Adding pod path:
  # C:\etc\kubernetes
  # I1023 23:44:11.775601    2468 file.go:68] Watching path
  # "C:\\etc\\kubernetes"
  # ...
  # E1023 23:44:31.794327    2468 file.go:182] Can't process manifest file
  # "C:\\etc\\kubernetes\\hns.psm1": C:\etc\kubernetes\hns.psm1: couldn't parse
  # as pod(yaml: line 10: did not find expected <document start>), please check
  # config file.
  #
  # Figure out how to change the directory that the kubelet monitors for new
  # pod manifests.

  # We configure the service to restart on failure, after 10s wait. We reset
  # the restart count to 0 each time, so we re-use our restart/10000 action on
  # each failure. Note it currently restarts even when explicitly stopped, you
  # have to delete the service entry to *really* kill it (e.g. `sc.exe delete
  # kubelet`). See issue #72900.
  if (Get-Process | Where-Object Name -eq "kubelet") {
    Log-Output -Fatal `
        "A kubelet process is already running, don't know what to do"
  }
  Log-Output "Creating kubelet service"
  sc.exe create kubelet binPath= "${env:NODE_DIR}\kubelet.exe ${kubelet_args}" start= demand
  sc.exe failure kubelet reset= 0 actions= restart/10000
  Log-Output "Starting kubelet service"
  sc.exe start kubelet

  Log-Output "Waiting 10 seconds for kubelet to stabilize"
  Start-Sleep 10

  if (Get-Process | Where-Object Name -eq "kube-proxy") {
    Log-Output -Fatal `
        "A kube-proxy process is already running, don't know what to do"
  }
  Log-Output "Creating kube-proxy service"
  sc.exe create kube-proxy binPath= "${env:NODE_DIR}\kube-proxy.exe ${kubeproxy_args}" start= demand
  sc.exe failure kube-proxy reset= 0 actions= restart/10000
  Log-Output "Starting kube-proxy service"
  sc.exe start kube-proxy

  # F1020 23:08:52.000083    9136 server.go:361] unable to load in-cluster
  # configuration, KUBERNETES_SERVICE_HOST and KUBERNETES_SERVICE_PORT must be
  # defined
  # TODO(pjh): still getting errors like these in kube-proxy log:
  # E1023 04:03:58.143449    4840 reflector.go:205] k8s.io/kubernetes/pkg/client/informers/informers_generated/internalversion/factory.go:129: Failed to list *core.Endpoints: Get https://35.239.84.171/api/v1/endpoints?limit=500&resourceVersion=0: dial tcp 35.239.84.171:443: connectex: A connection attempt failed because the connected party did not properly respond after a period of time, or established connection failed because connected host has failed to respond.
  # E1023 04:03:58.150266    4840 reflector.go:205] k8s.io/kubernetes/pkg/client/informers/informers_generated/internalversion/factory.go:129: Failed to list *core.Service: Get https://35.239.84.171/api/v1/services?limit=500&resourceVersion=0: dial tcp 35.239.84.171:443: connectex: A connection attempt failed because the connected party did not properly respond after a period of time, or established connection failed because connected host has failed to respond.

  Log_Todo ("verify that jobs are still running; print more details about " +
            "the background jobs.")
  Log-Output "$(Get-Service kube* | Out-String)"
  Verify_GceMetadataServerRouteIsPresent
  Log-Output "Kubernetes components started successfully"
}

# Runs 'kubectl get nodes'.
# TODO(pjh): run more verification commands.
function Verify-WorkerServices {
  Log-Output ("kubectl get nodes:`n" +
              "$(& ${env:NODE_DIR}\kubectl.exe get nodes | Out-String)")
  Verify_GceMetadataServerRouteIsPresent
  Log_Todo "run more verification commands."
}

# Export all public functions:
Export-ModuleMember -Function *-*
