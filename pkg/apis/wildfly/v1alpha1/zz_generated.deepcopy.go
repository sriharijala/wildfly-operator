// +build !ignore_autogenerated

// This file was autogenerated by deepcopy-gen. Do not edit it manually!

package v1alpha1

import (
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DataSourceSpec) DeepCopyInto(out *DataSourceSpec) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DataSourceSpec.
func (in *DataSourceSpec) DeepCopy() *DataSourceSpec {
	if in == nil {
		return nil
	}
	out := new(DataSourceSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *WildflyAppServer) DeepCopyInto(out *WildflyAppServer) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new WildflyAppServer.
func (in *WildflyAppServer) DeepCopy() *WildflyAppServer {
	if in == nil {
		return nil
	}
	out := new(WildflyAppServer)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *WildflyAppServer) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	} else {
		return nil
	}
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *WildflyAppServerList) DeepCopyInto(out *WildflyAppServerList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	out.ListMeta = in.ListMeta
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]WildflyAppServer, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new WildflyAppServerList.
func (in *WildflyAppServerList) DeepCopy() *WildflyAppServerList {
	if in == nil {
		return nil
	}
	out := new(WildflyAppServerList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *WildflyAppServerList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	} else {
		return nil
	}
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *WildflyAppServerSpec) DeepCopyInto(out *WildflyAppServerSpec) {
	*out = *in
	if in.DataSourceConfig != nil {
		in, out := &in.DataSourceConfig, &out.DataSourceConfig
		*out = make(map[string]DataSourceSpec, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new WildflyAppServerSpec.
func (in *WildflyAppServerSpec) DeepCopy() *WildflyAppServerSpec {
	if in == nil {
		return nil
	}
	out := new(WildflyAppServerSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *WildflyAppServerStatus) DeepCopyInto(out *WildflyAppServerStatus) {
	*out = *in
	if in.Nodes != nil {
		in, out := &in.Nodes, &out.Nodes
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.ExternalAddresses != nil {
		in, out := &in.ExternalAddresses, &out.ExternalAddresses
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new WildflyAppServerStatus.
func (in *WildflyAppServerStatus) DeepCopy() *WildflyAppServerStatus {
	if in == nil {
		return nil
	}
	out := new(WildflyAppServerStatus)
	in.DeepCopyInto(out)
	return out
}
