package(default_visibility = ["//visibility:public"])

<<<<<<< HEAD
licenses(["notice"])

=======
>>>>>>> 66f5f2bce071b09222a7a83d1f196f60c34cd224
load(
    "@io_bazel_rules_go//go:def.bzl",
    "go_library",
    "go_test",
)

go_test(
    name = "go_default_test",
    srcs = [
        "badconfig_test.go",
        "fstracker_test.go",
    ],
    library = ":go_default_library",
<<<<<<< HEAD
    tags = ["automanaged"],
=======
>>>>>>> 66f5f2bce071b09222a7a83d1f196f60c34cd224
    deps = [
        "//pkg/kubelet/kubeletconfig/util/files:go_default_library",
        "//pkg/kubelet/kubeletconfig/util/filesystem:go_default_library",
        "//pkg/kubelet/kubeletconfig/util/test:go_default_library",
    ],
)

go_library(
    name = "go_default_library",
    srcs = [
        "badconfig.go",
        "fstracker.go",
    ],
<<<<<<< HEAD
    tags = ["automanaged"],
=======
>>>>>>> 66f5f2bce071b09222a7a83d1f196f60c34cd224
    deps = [
        "//pkg/kubelet/kubeletconfig/util/files:go_default_library",
        "//pkg/kubelet/kubeletconfig/util/filesystem:go_default_library",
        "//pkg/kubelet/kubeletconfig/util/log:go_default_library",
    ],
)

filegroup(
    name = "package-srcs",
    srcs = glob(["**"]),
    tags = ["automanaged"],
    visibility = ["//visibility:private"],
)

filegroup(
    name = "all-srcs",
    srcs = [":package-srcs"],
    tags = ["automanaged"],
)
