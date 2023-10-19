package kube

import (
	"gotest.tools/v3/fs"
	"testing"
)

func TestExtractContainerIDFromCgroupPath(t *testing.T) {
	type args struct {
		cgroupPath string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "case0",
			args: args{
				cgroupPath: "/kubepods.slice/kubepods-besteffort.slice/kubepods-besteffort-pod2c5a3044_0ccb_42f9_877f_833195e99229.slice/cri-containerd-284be7d2d89ff0e2c2be78be0885f659a487aff485fde68c130521e6b0126b10.scope",
			},
			want: "284be7d2d89ff0e2c2be78be0885f659a487aff485fde68c130521e6b0126b10",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ExtractContainerIDFromCgroupPath(tt.args.cgroupPath); got != tt.want {
				t.Errorf("ExtractContainerIDFromCgroupPath() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseCgroupFile(t *testing.T) {
	type args struct {
		content string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "v2",
			args: args{
				content: `0::/kubepods.slice/kubepods-pod268298b3_5747_4150_834a_de56008a6954.slice/cri-containerd-7adcc1db52494dee0cd74afcb34c2492af7a150f667c26d03b8bff8ab242db29.scope`,
			},
			want: "/kubepods.slice/kubepods-pod268298b3_5747_4150_834a_de56008a6954.slice/cri-containerd-7adcc1db52494dee0cd74afcb34c2492af7a150f667c26d03b8bff8ab242db29.scope",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			file := fs.NewFile(t, "cgroup", fs.WithContent(tt.args.content))
			got, err := ParseCgroupFile(file.Path())
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseCgroupFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ParseCgroupFile() got = %v, want %v", got, tt.want)
			}
		})
	}
}
