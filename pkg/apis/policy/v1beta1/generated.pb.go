/*
Copyright 2017 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Code generated by protoc-gen-gogo.
// source: k8s.io/kubernetes/pkg/apis/policy/v1beta1/generated.proto
// DO NOT EDIT!

/*
	Package v1beta1 is a generated protocol buffer package.

	It is generated from these files:
		k8s.io/kubernetes/pkg/apis/policy/v1beta1/generated.proto

	It has these top-level messages:
		Eviction
		PodDisruptionBudget
		PodDisruptionBudgetList
		PodDisruptionBudgetSpec
		PodDisruptionBudgetStatus
*/
package v1beta1

import proto "github.com/gogo/protobuf/proto"
import fmt "fmt"
import math "math"

import k8s_io_apimachinery_pkg_apis_meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

import strings "strings"
import reflect "reflect"
import github_com_gogo_protobuf_sortkeys "github.com/gogo/protobuf/sortkeys"

import io "io"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
const _ = proto.GoGoProtoPackageIsVersion1

func (m *Eviction) Reset()                    { *m = Eviction{} }
func (*Eviction) ProtoMessage()               {}
func (*Eviction) Descriptor() ([]byte, []int) { return fileDescriptorGenerated, []int{0} }

func (m *PodDisruptionBudget) Reset()                    { *m = PodDisruptionBudget{} }
func (*PodDisruptionBudget) ProtoMessage()               {}
func (*PodDisruptionBudget) Descriptor() ([]byte, []int) { return fileDescriptorGenerated, []int{1} }

func (m *PodDisruptionBudgetList) Reset()                    { *m = PodDisruptionBudgetList{} }
func (*PodDisruptionBudgetList) ProtoMessage()               {}
func (*PodDisruptionBudgetList) Descriptor() ([]byte, []int) { return fileDescriptorGenerated, []int{2} }

func (m *PodDisruptionBudgetSpec) Reset()                    { *m = PodDisruptionBudgetSpec{} }
func (*PodDisruptionBudgetSpec) ProtoMessage()               {}
func (*PodDisruptionBudgetSpec) Descriptor() ([]byte, []int) { return fileDescriptorGenerated, []int{3} }

func (m *PodDisruptionBudgetStatus) Reset()      { *m = PodDisruptionBudgetStatus{} }
func (*PodDisruptionBudgetStatus) ProtoMessage() {}
func (*PodDisruptionBudgetStatus) Descriptor() ([]byte, []int) {
	return fileDescriptorGenerated, []int{4}
}

func init() {
	proto.RegisterType((*Eviction)(nil), "k8s.io.kubernetes.pkg.apis.policy.v1beta1.Eviction")
	proto.RegisterType((*PodDisruptionBudget)(nil), "k8s.io.kubernetes.pkg.apis.policy.v1beta1.PodDisruptionBudget")
	proto.RegisterType((*PodDisruptionBudgetList)(nil), "k8s.io.kubernetes.pkg.apis.policy.v1beta1.PodDisruptionBudgetList")
	proto.RegisterType((*PodDisruptionBudgetSpec)(nil), "k8s.io.kubernetes.pkg.apis.policy.v1beta1.PodDisruptionBudgetSpec")
	proto.RegisterType((*PodDisruptionBudgetStatus)(nil), "k8s.io.kubernetes.pkg.apis.policy.v1beta1.PodDisruptionBudgetStatus")
}
func (m *Eviction) Marshal() (data []byte, err error) {
	size := m.Size()
	data = make([]byte, size)
	n, err := m.MarshalTo(data)
	if err != nil {
		return nil, err
	}
	return data[:n], nil
}

func (m *Eviction) MarshalTo(data []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	data[i] = 0xa
	i++
	i = encodeVarintGenerated(data, i, uint64(m.ObjectMeta.Size()))
	n1, err := m.ObjectMeta.MarshalTo(data[i:])
	if err != nil {
		return 0, err
	}
	i += n1
	if m.DeleteOptions != nil {
		data[i] = 0x12
		i++
		i = encodeVarintGenerated(data, i, uint64(m.DeleteOptions.Size()))
		n2, err := m.DeleteOptions.MarshalTo(data[i:])
		if err != nil {
			return 0, err
		}
		i += n2
	}
	return i, nil
}

func (m *PodDisruptionBudget) Marshal() (data []byte, err error) {
	size := m.Size()
	data = make([]byte, size)
	n, err := m.MarshalTo(data)
	if err != nil {
		return nil, err
	}
	return data[:n], nil
}

func (m *PodDisruptionBudget) MarshalTo(data []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	data[i] = 0xa
	i++
	i = encodeVarintGenerated(data, i, uint64(m.ObjectMeta.Size()))
	n3, err := m.ObjectMeta.MarshalTo(data[i:])
	if err != nil {
		return 0, err
	}
	i += n3
	data[i] = 0x12
	i++
	i = encodeVarintGenerated(data, i, uint64(m.Spec.Size()))
	n4, err := m.Spec.MarshalTo(data[i:])
	if err != nil {
		return 0, err
	}
	i += n4
	data[i] = 0x1a
	i++
	i = encodeVarintGenerated(data, i, uint64(m.Status.Size()))
	n5, err := m.Status.MarshalTo(data[i:])
	if err != nil {
		return 0, err
	}
	i += n5
	return i, nil
}

func (m *PodDisruptionBudgetList) Marshal() (data []byte, err error) {
	size := m.Size()
	data = make([]byte, size)
	n, err := m.MarshalTo(data)
	if err != nil {
		return nil, err
	}
	return data[:n], nil
}

func (m *PodDisruptionBudgetList) MarshalTo(data []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	data[i] = 0xa
	i++
	i = encodeVarintGenerated(data, i, uint64(m.ListMeta.Size()))
	n6, err := m.ListMeta.MarshalTo(data[i:])
	if err != nil {
		return 0, err
	}
	i += n6
	if len(m.Items) > 0 {
		for _, msg := range m.Items {
			data[i] = 0x12
			i++
			i = encodeVarintGenerated(data, i, uint64(msg.Size()))
			n, err := msg.MarshalTo(data[i:])
			if err != nil {
				return 0, err
			}
			i += n
		}
	}
	return i, nil
}

func (m *PodDisruptionBudgetSpec) Marshal() (data []byte, err error) {
	size := m.Size()
	data = make([]byte, size)
	n, err := m.MarshalTo(data)
	if err != nil {
		return nil, err
	}
	return data[:n], nil
}

func (m *PodDisruptionBudgetSpec) MarshalTo(data []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	data[i] = 0xa
	i++
	i = encodeVarintGenerated(data, i, uint64(m.MinAvailable.Size()))
	n7, err := m.MinAvailable.MarshalTo(data[i:])
	if err != nil {
		return 0, err
	}
	i += n7
	if m.Selector != nil {
		data[i] = 0x12
		i++
		i = encodeVarintGenerated(data, i, uint64(m.Selector.Size()))
		n8, err := m.Selector.MarshalTo(data[i:])
		if err != nil {
			return 0, err
		}
		i += n8
	}
	return i, nil
}

func (m *PodDisruptionBudgetStatus) Marshal() (data []byte, err error) {
	size := m.Size()
	data = make([]byte, size)
	n, err := m.MarshalTo(data)
	if err != nil {
		return nil, err
	}
	return data[:n], nil
}

func (m *PodDisruptionBudgetStatus) MarshalTo(data []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	data[i] = 0x8
	i++
	i = encodeVarintGenerated(data, i, uint64(m.ObservedGeneration))
	if len(m.DisruptedPods) > 0 {
		for k := range m.DisruptedPods {
			data[i] = 0x12
			i++
			v := m.DisruptedPods[k]
			msgSize := (&v).Size()
			mapSize := 1 + len(k) + sovGenerated(uint64(len(k))) + 1 + msgSize + sovGenerated(uint64(msgSize))
			i = encodeVarintGenerated(data, i, uint64(mapSize))
			data[i] = 0xa
			i++
			i = encodeVarintGenerated(data, i, uint64(len(k)))
			i += copy(data[i:], k)
			data[i] = 0x12
			i++
			i = encodeVarintGenerated(data, i, uint64((&v).Size()))
			n9, err := (&v).MarshalTo(data[i:])
			if err != nil {
				return 0, err
			}
			i += n9
		}
	}
	data[i] = 0x18
	i++
	i = encodeVarintGenerated(data, i, uint64(m.PodDisruptionsAllowed))
	data[i] = 0x20
	i++
	i = encodeVarintGenerated(data, i, uint64(m.CurrentHealthy))
	data[i] = 0x28
	i++
	i = encodeVarintGenerated(data, i, uint64(m.DesiredHealthy))
	data[i] = 0x30
	i++
	i = encodeVarintGenerated(data, i, uint64(m.ExpectedPods))
	return i, nil
}

func encodeFixed64Generated(data []byte, offset int, v uint64) int {
	data[offset] = uint8(v)
	data[offset+1] = uint8(v >> 8)
	data[offset+2] = uint8(v >> 16)
	data[offset+3] = uint8(v >> 24)
	data[offset+4] = uint8(v >> 32)
	data[offset+5] = uint8(v >> 40)
	data[offset+6] = uint8(v >> 48)
	data[offset+7] = uint8(v >> 56)
	return offset + 8
}
func encodeFixed32Generated(data []byte, offset int, v uint32) int {
	data[offset] = uint8(v)
	data[offset+1] = uint8(v >> 8)
	data[offset+2] = uint8(v >> 16)
	data[offset+3] = uint8(v >> 24)
	return offset + 4
}
func encodeVarintGenerated(data []byte, offset int, v uint64) int {
	for v >= 1<<7 {
		data[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	data[offset] = uint8(v)
	return offset + 1
}
func (m *Eviction) Size() (n int) {
	var l int
	_ = l
	l = m.ObjectMeta.Size()
	n += 1 + l + sovGenerated(uint64(l))
	if m.DeleteOptions != nil {
		l = m.DeleteOptions.Size()
		n += 1 + l + sovGenerated(uint64(l))
	}
	return n
}

func (m *PodDisruptionBudget) Size() (n int) {
	var l int
	_ = l
	l = m.ObjectMeta.Size()
	n += 1 + l + sovGenerated(uint64(l))
	l = m.Spec.Size()
	n += 1 + l + sovGenerated(uint64(l))
	l = m.Status.Size()
	n += 1 + l + sovGenerated(uint64(l))
	return n
}

func (m *PodDisruptionBudgetList) Size() (n int) {
	var l int
	_ = l
	l = m.ListMeta.Size()
	n += 1 + l + sovGenerated(uint64(l))
	if len(m.Items) > 0 {
		for _, e := range m.Items {
			l = e.Size()
			n += 1 + l + sovGenerated(uint64(l))
		}
	}
	return n
}

func (m *PodDisruptionBudgetSpec) Size() (n int) {
	var l int
	_ = l
	l = m.MinAvailable.Size()
	n += 1 + l + sovGenerated(uint64(l))
	if m.Selector != nil {
		l = m.Selector.Size()
		n += 1 + l + sovGenerated(uint64(l))
	}
	return n
}

func (m *PodDisruptionBudgetStatus) Size() (n int) {
	var l int
	_ = l
	n += 1 + sovGenerated(uint64(m.ObservedGeneration))
	if len(m.DisruptedPods) > 0 {
		for k, v := range m.DisruptedPods {
			_ = k
			_ = v
			l = v.Size()
			mapEntrySize := 1 + len(k) + sovGenerated(uint64(len(k))) + 1 + l + sovGenerated(uint64(l))
			n += mapEntrySize + 1 + sovGenerated(uint64(mapEntrySize))
		}
	}
	n += 1 + sovGenerated(uint64(m.PodDisruptionsAllowed))
	n += 1 + sovGenerated(uint64(m.CurrentHealthy))
	n += 1 + sovGenerated(uint64(m.DesiredHealthy))
	n += 1 + sovGenerated(uint64(m.ExpectedPods))
	return n
}

func sovGenerated(x uint64) (n int) {
	for {
		n++
		x >>= 7
		if x == 0 {
			break
		}
	}
	return n
}
func sozGenerated(x uint64) (n int) {
	return sovGenerated(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (this *Eviction) String() string {
	if this == nil {
		return "nil"
	}
	s := strings.Join([]string{`&Eviction{`,
		`ObjectMeta:` + strings.Replace(strings.Replace(this.ObjectMeta.String(), "ObjectMeta", "k8s_io_apimachinery_pkg_apis_meta_v1.ObjectMeta", 1), `&`, ``, 1) + `,`,
		`DeleteOptions:` + strings.Replace(fmt.Sprintf("%v", this.DeleteOptions), "DeleteOptions", "k8s_io_apimachinery_pkg_apis_meta_v1.DeleteOptions", 1) + `,`,
		`}`,
	}, "")
	return s
}
func (this *PodDisruptionBudget) String() string {
	if this == nil {
		return "nil"
	}
	s := strings.Join([]string{`&PodDisruptionBudget{`,
		`ObjectMeta:` + strings.Replace(strings.Replace(this.ObjectMeta.String(), "ObjectMeta", "k8s_io_apimachinery_pkg_apis_meta_v1.ObjectMeta", 1), `&`, ``, 1) + `,`,
		`Spec:` + strings.Replace(strings.Replace(this.Spec.String(), "PodDisruptionBudgetSpec", "PodDisruptionBudgetSpec", 1), `&`, ``, 1) + `,`,
		`Status:` + strings.Replace(strings.Replace(this.Status.String(), "PodDisruptionBudgetStatus", "PodDisruptionBudgetStatus", 1), `&`, ``, 1) + `,`,
		`}`,
	}, "")
	return s
}
func (this *PodDisruptionBudgetList) String() string {
	if this == nil {
		return "nil"
	}
	s := strings.Join([]string{`&PodDisruptionBudgetList{`,
		`ListMeta:` + strings.Replace(strings.Replace(this.ListMeta.String(), "ListMeta", "k8s_io_apimachinery_pkg_apis_meta_v1.ListMeta", 1), `&`, ``, 1) + `,`,
		`Items:` + strings.Replace(strings.Replace(fmt.Sprintf("%v", this.Items), "PodDisruptionBudget", "PodDisruptionBudget", 1), `&`, ``, 1) + `,`,
		`}`,
	}, "")
	return s
}
func (this *PodDisruptionBudgetSpec) String() string {
	if this == nil {
		return "nil"
	}
	s := strings.Join([]string{`&PodDisruptionBudgetSpec{`,
		`MinAvailable:` + strings.Replace(strings.Replace(this.MinAvailable.String(), "IntOrString", "k8s_io_kubernetes_pkg_util_intstr.IntOrString", 1), `&`, ``, 1) + `,`,
		`Selector:` + strings.Replace(fmt.Sprintf("%v", this.Selector), "LabelSelector", "k8s_io_apimachinery_pkg_apis_meta_v1.LabelSelector", 1) + `,`,
		`}`,
	}, "")
	return s
}
func (this *PodDisruptionBudgetStatus) String() string {
	if this == nil {
		return "nil"
	}
	keysForDisruptedPods := make([]string, 0, len(this.DisruptedPods))
	for k := range this.DisruptedPods {
		keysForDisruptedPods = append(keysForDisruptedPods, k)
	}
	github_com_gogo_protobuf_sortkeys.Strings(keysForDisruptedPods)
	mapStringForDisruptedPods := "map[string]k8s_io_apimachinery_pkg_apis_meta_v1.Time{"
	for _, k := range keysForDisruptedPods {
		mapStringForDisruptedPods += fmt.Sprintf("%v: %v,", k, this.DisruptedPods[k])
	}
	mapStringForDisruptedPods += "}"
	s := strings.Join([]string{`&PodDisruptionBudgetStatus{`,
		`ObservedGeneration:` + fmt.Sprintf("%v", this.ObservedGeneration) + `,`,
		`DisruptedPods:` + mapStringForDisruptedPods + `,`,
		`PodDisruptionsAllowed:` + fmt.Sprintf("%v", this.PodDisruptionsAllowed) + `,`,
		`CurrentHealthy:` + fmt.Sprintf("%v", this.CurrentHealthy) + `,`,
		`DesiredHealthy:` + fmt.Sprintf("%v", this.DesiredHealthy) + `,`,
		`ExpectedPods:` + fmt.Sprintf("%v", this.ExpectedPods) + `,`,
		`}`,
	}, "")
	return s
}
func valueToStringGenerated(v interface{}) string {
	rv := reflect.ValueOf(v)
	if rv.IsNil() {
		return "nil"
	}
	pv := reflect.Indirect(rv).Interface()
	return fmt.Sprintf("*%v", pv)
}
func (m *Eviction) Unmarshal(data []byte) error {
	l := len(data)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowGenerated
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := data[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: Eviction: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: Eviction: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field ObjectMeta", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenerated
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := data[iNdEx]
				iNdEx++
				msglen |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthGenerated
			}
			postIndex := iNdEx + msglen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.ObjectMeta.Unmarshal(data[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field DeleteOptions", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenerated
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := data[iNdEx]
				iNdEx++
				msglen |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthGenerated
			}
			postIndex := iNdEx + msglen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if m.DeleteOptions == nil {
				m.DeleteOptions = &k8s_io_apimachinery_pkg_apis_meta_v1.DeleteOptions{}
			}
			if err := m.DeleteOptions.Unmarshal(data[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipGenerated(data[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthGenerated
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *PodDisruptionBudget) Unmarshal(data []byte) error {
	l := len(data)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowGenerated
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := data[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: PodDisruptionBudget: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: PodDisruptionBudget: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field ObjectMeta", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenerated
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := data[iNdEx]
				iNdEx++
				msglen |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthGenerated
			}
			postIndex := iNdEx + msglen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.ObjectMeta.Unmarshal(data[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Spec", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenerated
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := data[iNdEx]
				iNdEx++
				msglen |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthGenerated
			}
			postIndex := iNdEx + msglen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.Spec.Unmarshal(data[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Status", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenerated
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := data[iNdEx]
				iNdEx++
				msglen |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthGenerated
			}
			postIndex := iNdEx + msglen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.Status.Unmarshal(data[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipGenerated(data[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthGenerated
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *PodDisruptionBudgetList) Unmarshal(data []byte) error {
	l := len(data)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowGenerated
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := data[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: PodDisruptionBudgetList: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: PodDisruptionBudgetList: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field ListMeta", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenerated
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := data[iNdEx]
				iNdEx++
				msglen |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthGenerated
			}
			postIndex := iNdEx + msglen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.ListMeta.Unmarshal(data[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Items", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenerated
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := data[iNdEx]
				iNdEx++
				msglen |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthGenerated
			}
			postIndex := iNdEx + msglen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Items = append(m.Items, PodDisruptionBudget{})
			if err := m.Items[len(m.Items)-1].Unmarshal(data[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipGenerated(data[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthGenerated
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *PodDisruptionBudgetSpec) Unmarshal(data []byte) error {
	l := len(data)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowGenerated
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := data[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: PodDisruptionBudgetSpec: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: PodDisruptionBudgetSpec: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field MinAvailable", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenerated
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := data[iNdEx]
				iNdEx++
				msglen |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthGenerated
			}
			postIndex := iNdEx + msglen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.MinAvailable.Unmarshal(data[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Selector", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenerated
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := data[iNdEx]
				iNdEx++
				msglen |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthGenerated
			}
			postIndex := iNdEx + msglen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if m.Selector == nil {
				m.Selector = &k8s_io_apimachinery_pkg_apis_meta_v1.LabelSelector{}
			}
			if err := m.Selector.Unmarshal(data[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipGenerated(data[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthGenerated
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *PodDisruptionBudgetStatus) Unmarshal(data []byte) error {
	l := len(data)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowGenerated
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := data[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: PodDisruptionBudgetStatus: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: PodDisruptionBudgetStatus: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field ObservedGeneration", wireType)
			}
			m.ObservedGeneration = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenerated
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := data[iNdEx]
				iNdEx++
				m.ObservedGeneration |= (int64(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field DisruptedPods", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenerated
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := data[iNdEx]
				iNdEx++
				msglen |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthGenerated
			}
			postIndex := iNdEx + msglen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			var keykey uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenerated
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := data[iNdEx]
				iNdEx++
				keykey |= (uint64(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			var stringLenmapkey uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenerated
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := data[iNdEx]
				iNdEx++
				stringLenmapkey |= (uint64(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLenmapkey := int(stringLenmapkey)
			if intStringLenmapkey < 0 {
				return ErrInvalidLengthGenerated
			}
			postStringIndexmapkey := iNdEx + intStringLenmapkey
			if postStringIndexmapkey > l {
				return io.ErrUnexpectedEOF
			}
			mapkey := string(data[iNdEx:postStringIndexmapkey])
			iNdEx = postStringIndexmapkey
			var valuekey uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenerated
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := data[iNdEx]
				iNdEx++
				valuekey |= (uint64(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			var mapmsglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenerated
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := data[iNdEx]
				iNdEx++
				mapmsglen |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if mapmsglen < 0 {
				return ErrInvalidLengthGenerated
			}
			postmsgIndex := iNdEx + mapmsglen
			if mapmsglen < 0 {
				return ErrInvalidLengthGenerated
			}
			if postmsgIndex > l {
				return io.ErrUnexpectedEOF
			}
			mapvalue := &k8s_io_apimachinery_pkg_apis_meta_v1.Time{}
			if err := mapvalue.Unmarshal(data[iNdEx:postmsgIndex]); err != nil {
				return err
			}
			iNdEx = postmsgIndex
			if m.DisruptedPods == nil {
				m.DisruptedPods = make(map[string]k8s_io_apimachinery_pkg_apis_meta_v1.Time)
			}
			m.DisruptedPods[mapkey] = *mapvalue
			iNdEx = postIndex
		case 3:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field PodDisruptionsAllowed", wireType)
			}
			m.PodDisruptionsAllowed = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenerated
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := data[iNdEx]
				iNdEx++
				m.PodDisruptionsAllowed |= (int32(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 4:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field CurrentHealthy", wireType)
			}
			m.CurrentHealthy = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenerated
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := data[iNdEx]
				iNdEx++
				m.CurrentHealthy |= (int32(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 5:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field DesiredHealthy", wireType)
			}
			m.DesiredHealthy = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenerated
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := data[iNdEx]
				iNdEx++
				m.DesiredHealthy |= (int32(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 6:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field ExpectedPods", wireType)
			}
			m.ExpectedPods = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenerated
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := data[iNdEx]
				iNdEx++
				m.ExpectedPods |= (int32(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		default:
			iNdEx = preIndex
			skippy, err := skipGenerated(data[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthGenerated
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func skipGenerated(data []byte) (n int, err error) {
	l := len(data)
	iNdEx := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowGenerated
			}
			if iNdEx >= l {
				return 0, io.ErrUnexpectedEOF
			}
			b := data[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		wireType := int(wire & 0x7)
		switch wireType {
		case 0:
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowGenerated
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				iNdEx++
				if data[iNdEx-1] < 0x80 {
					break
				}
			}
			return iNdEx, nil
		case 1:
			iNdEx += 8
			return iNdEx, nil
		case 2:
			var length int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowGenerated
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				b := data[iNdEx]
				iNdEx++
				length |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			iNdEx += length
			if length < 0 {
				return 0, ErrInvalidLengthGenerated
			}
			return iNdEx, nil
		case 3:
			for {
				var innerWire uint64
				var start int = iNdEx
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return 0, ErrIntOverflowGenerated
					}
					if iNdEx >= l {
						return 0, io.ErrUnexpectedEOF
					}
					b := data[iNdEx]
					iNdEx++
					innerWire |= (uint64(b) & 0x7F) << shift
					if b < 0x80 {
						break
					}
				}
				innerWireType := int(innerWire & 0x7)
				if innerWireType == 4 {
					break
				}
				next, err := skipGenerated(data[start:])
				if err != nil {
					return 0, err
				}
				iNdEx = start + next
			}
			return iNdEx, nil
		case 4:
			return iNdEx, nil
		case 5:
			iNdEx += 4
			return iNdEx, nil
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
	}
	panic("unreachable")
}

var (
	ErrInvalidLengthGenerated = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowGenerated   = fmt.Errorf("proto: integer overflow")
)

var fileDescriptorGenerated = []byte{
	// 776 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x09, 0x6e, 0x88, 0x02, 0xff, 0xbc, 0x94, 0xcf, 0x6f, 0xe3, 0x44,
	0x14, 0xc7, 0xe3, 0x26, 0x29, 0x61, 0x9a, 0x54, 0x65, 0xa0, 0x10, 0x22, 0xe1, 0xa2, 0x9c, 0x5a,
	0x04, 0x63, 0x52, 0x10, 0x2a, 0x1c, 0x2a, 0x6a, 0x52, 0x41, 0x51, 0xab, 0x54, 0x2e, 0x12, 0x12,
	0x02, 0x89, 0xb1, 0xfd, 0x70, 0x86, 0xf8, 0x97, 0x66, 0xc6, 0x81, 0xdc, 0xf8, 0x13, 0x38, 0xec,
	0x1f, 0x55, 0x69, 0x2f, 0x3d, 0xae, 0x56, 0xab, 0x6a, 0x9b, 0xfd, 0x1b, 0xf6, 0xbe, 0xb2, 0x3d,
	0x49, 0xe3, 0x24, 0xdd, 0xcd, 0xaa, 0xab, 0xbd, 0x79, 0x66, 0xde, 0xe7, 0xfb, 0x7d, 0xef, 0xcd,
	0x1b, 0xa3, 0x6f, 0x07, 0x07, 0x82, 0xb0, 0xc8, 0x18, 0x24, 0x36, 0xf0, 0x10, 0x24, 0x08, 0x23,
	0x1e, 0x78, 0x06, 0x8d, 0x99, 0x30, 0xe2, 0xc8, 0x67, 0xce, 0xc8, 0x18, 0x76, 0x6c, 0x90, 0xb4,
	0x63, 0x78, 0x10, 0x02, 0xa7, 0x12, 0x5c, 0x12, 0xf3, 0x48, 0x46, 0x78, 0x2f, 0x47, 0xc9, 0x2d,
	0x4a, 0xe2, 0x81, 0x47, 0x52, 0x94, 0xe4, 0x28, 0x51, 0x68, 0xeb, 0x0b, 0x8f, 0xc9, 0x7e, 0x62,
	0x13, 0x27, 0x0a, 0x0c, 0x2f, 0xf2, 0x22, 0x23, 0x53, 0xb0, 0x93, 0xbf, 0xb2, 0x55, 0xb6, 0xc8,
	0xbe, 0x72, 0xe5, 0xd6, 0xd7, 0x2a, 0x29, 0x1a, 0xb3, 0x80, 0x3a, 0x7d, 0x16, 0x02, 0x1f, 0x4d,
	0xd2, 0x32, 0x38, 0x88, 0x28, 0xe1, 0x0e, 0xcc, 0xe7, 0xf3, 0x52, 0x4a, 0x18, 0x01, 0x48, 0x6a,
	0x0c, 0x17, 0xaa, 0x68, 0x19, 0x77, 0x51, 0x3c, 0x09, 0x25, 0x0b, 0x16, 0x6d, 0xbe, 0x79, 0x15,
	0x20, 0x9c, 0x3e, 0x04, 0x74, 0x81, 0xfb, 0xfc, 0xce, 0x4e, 0x2f, 0x4b, 0xab, 0xb3, 0x3c, 0x3a,
	0x91, 0xcc, 0x37, 0x58, 0x28, 0x85, 0xe4, 0xf3, 0x48, 0xfb, 0xb1, 0x86, 0x6a, 0xc7, 0x43, 0xe6,
	0x48, 0x16, 0x85, 0xf8, 0x4f, 0x54, 0x4b, 0x2b, 0x76, 0xa9, 0xa4, 0x4d, 0xed, 0x53, 0x6d, 0x77,
	0x63, 0xff, 0x4b, 0xa2, 0xee, 0x6b, 0x36, 0xf1, 0xdb, 0x1b, 0x4b, 0xa3, 0xc9, 0xb0, 0x43, 0x7a,
	0xf6, 0xdf, 0xe0, 0xc8, 0x33, 0x90, 0xd4, 0xc4, 0x97, 0xd7, 0x3b, 0xa5, 0xf1, 0xf5, 0x0e, 0xba,
	0xdd, 0xb3, 0xa6, 0xaa, 0xd8, 0x47, 0x0d, 0x17, 0x7c, 0x90, 0xd0, 0x8b, 0x53, 0x47, 0xd1, 0x5c,
	0xcb, 0x6c, 0xbe, 0x5a, 0xcd, 0xa6, 0x3b, 0x8b, 0x9a, 0xef, 0x8d, 0xaf, 0x77, 0x1a, 0x85, 0x2d,
	0xab, 0x28, 0xde, 0x7e, 0xb8, 0x86, 0xde, 0x3f, 0x8f, 0xdc, 0x2e, 0x13, 0x3c, 0xc9, 0xb6, 0xcc,
	0xc4, 0xf5, 0x40, 0xbe, 0x85, 0x3a, 0x5d, 0x54, 0x11, 0x31, 0x38, 0xaa, 0x3c, 0x93, 0xac, 0x3c,
	0xf5, 0x64, 0x49, 0xbe, 0x17, 0x31, 0x38, 0x66, 0x5d, 0xf9, 0x55, 0xd2, 0x95, 0x95, 0xa9, 0x63,
	0x1f, 0xad, 0x0b, 0x49, 0x65, 0x22, 0x9a, 0xe5, 0xcc, 0xa7, 0x7b, 0x4f, 0x9f, 0x4c, 0xcb, 0xdc,
	0x54, 0x4e, 0xeb, 0xf9, 0xda, 0x52, 0x1e, 0xed, 0x27, 0x1a, 0xfa, 0x68, 0x09, 0x75, 0xca, 0x84,
	0xc4, 0xbf, 0x2f, 0x74, 0x94, 0xac, 0xd6, 0xd1, 0x94, 0xce, 0xfa, 0xb9, 0xa5, 0x5c, 0x6b, 0x93,
	0x9d, 0x99, 0x6e, 0x3a, 0xa8, 0xca, 0x24, 0x04, 0xe9, 0xb4, 0x94, 0x77, 0x37, 0xf6, 0x0f, 0xef,
	0x57, 0xa6, 0xd9, 0x50, 0x56, 0xd5, 0x93, 0x54, 0xd4, 0xca, 0xb5, 0xdb, 0xe3, 0xe5, 0xe5, 0xa5,
	0xed, 0xc6, 0x7d, 0x54, 0x0f, 0x58, 0x78, 0x34, 0xa4, 0xcc, 0xa7, 0xb6, 0x0f, 0xf3, 0x25, 0xce,
	0xe5, 0x91, 0xbe, 0x37, 0x92, 0xbf, 0x37, 0x72, 0x12, 0xca, 0x1e, 0xbf, 0x90, 0x9c, 0x85, 0x9e,
	0xf9, 0x81, 0xf2, 0xad, 0x9f, 0xcd, 0x68, 0x59, 0x05, 0x65, 0xfc, 0x07, 0xaa, 0x09, 0xf0, 0xc1,
	0x91, 0x11, 0x7f, 0xbd, 0xb7, 0x71, 0x4a, 0x6d, 0xf0, 0x2f, 0x14, 0x6a, 0xd6, 0xd3, 0x4e, 0x4e,
	0x56, 0xd6, 0x54, 0xb2, 0xfd, 0xbc, 0x82, 0x3e, 0xbe, 0xf3, 0xe6, 0xf1, 0xcf, 0x08, 0x47, 0xb6,
	0x00, 0x3e, 0x04, 0xf7, 0xc7, 0xfc, 0x3f, 0xc1, 0xa2, 0x30, 0x2b, 0xb6, 0x6c, 0xb6, 0x54, 0xf2,
	0xb8, 0xb7, 0x10, 0x61, 0x2d, 0xa1, 0xf0, 0x03, 0x0d, 0x35, 0xdc, 0xdc, 0x06, 0xdc, 0xf3, 0xc8,
	0x9d, 0x5c, 0xde, 0xaf, 0x6f, 0x62, 0x46, 0x49, 0x77, 0x56, 0xf9, 0x38, 0x94, 0x7c, 0x64, 0x6e,
	0xab, 0x04, 0x1b, 0x85, 0x33, 0xab, 0x98, 0x04, 0x3e, 0x43, 0xd8, 0x9d, 0x4a, 0x8a, 0x23, 0xdf,
	0x8f, 0xfe, 0x01, 0x37, 0x7b, 0x3e, 0x55, 0xf3, 0x13, 0xa5, 0xb0, 0x5d, 0xf0, 0x9d, 0x04, 0x59,
	0x4b, 0x40, 0x7c, 0x88, 0x36, 0x9d, 0x84, 0x73, 0x08, 0xe5, 0x4f, 0x40, 0x7d, 0xd9, 0x1f, 0x35,
	0x2b, 0x99, 0xd4, 0x87, 0x4a, 0x6a, 0xf3, 0x87, 0xc2, 0xa9, 0x35, 0x17, 0x9d, 0xf2, 0x2e, 0x08,
	0xc6, 0xc1, 0x9d, 0xf0, 0xd5, 0x22, 0xdf, 0x2d, 0x9c, 0x5a, 0x73, 0xd1, 0xf8, 0x00, 0xd5, 0xe1,
	0xdf, 0x18, 0x9c, 0x49, 0x8f, 0xd7, 0x33, 0x7a, 0x3a, 0x68, 0xc7, 0x33, 0x67, 0x56, 0x21, 0xb2,
	0xe5, 0x23, 0xbc, 0xd8, 0x44, 0xbc, 0x85, 0xca, 0x03, 0x18, 0x65, 0x57, 0xfe, 0xae, 0x95, 0x7e,
	0xe2, 0xef, 0x51, 0x75, 0x48, 0xfd, 0x04, 0xd4, 0x34, 0x7e, 0xb6, 0xda, 0x34, 0xfe, 0xc2, 0x02,
	0xb0, 0x72, 0xf0, 0xbb, 0xb5, 0x03, 0xcd, 0xdc, 0xbb, 0xbc, 0xd1, 0x4b, 0x57, 0x37, 0x7a, 0xe9,
	0xd1, 0x8d, 0x5e, 0xfa, 0x6f, 0xac, 0x6b, 0x97, 0x63, 0x5d, 0xbb, 0x1a, 0xeb, 0xda, 0xd3, 0xb1,
	0xae, 0xfd, 0xff, 0x4c, 0x2f, 0xfd, 0xf6, 0x8e, 0xba, 0xf4, 0x17, 0x01, 0x00, 0x00, 0xff, 0xff,
	0x02, 0xd3, 0x1e, 0x38, 0x5d, 0x08, 0x00, 0x00,
}
