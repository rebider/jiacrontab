// +build !windows

package kproc

import (
	"context"
	"github.com/iwannay/log"
	"os"
	"os/exec"
	"os/user"
	"strconv"
	"syscall"
)

func CommandContext(ctx context.Context, name string, arg ...string) *KCmd {
	cmd := exec.CommandContext(ctx, name, arg...)
	cmd.SysProcAttr = &syscall.SysProcAttr{}
	cmd.SysProcAttr.Setsid = true
	return &KCmd{
		ctx:  ctx,
		Cmd:  cmd,
		done: make(chan struct{}),
	}
}

func (k *KCmd) SetUser(username string) {
	u, err := user.Lookup(username)
	if err == nil {
		log.Infof("uid=%s,gid=%s", u.Uid, u.Gid)
		uid, _ := strconv.Atoi(u.Uid)
		gid, _ := strconv.Atoi(u.Gid)

		k.SysProcAttr = &syscall.SysProcAttr{}
		k.SysProcAttr.Credential = &syscall.Credential{Uid: uint32(uid), Gid: uint32(gid)}
	}
}

func (k *KCmd) SetEnv(env []string) {
	k.Cmd.Env = env
}

func (k *KCmd) SetDir(dir string) {
	k.Cmd.Dir = dir
}

func (k *KCmd) KillAll() {

	select {
	case k.done <- struct{}{}:
	default:
	}

	if k.Process == nil {
		return
	}

	group, err := os.FindProcess(-k.Process.Pid)
	if err == nil {
		group.Signal(syscall.SIGKILL)
	}
}

func (k *KCmd) Wait() error {
	defer k.KillAll()
	go func() {
		select {
		case <-k.ctx.Done():
			k.KillAll()
		case <-k.done:
		}
	}()
	return k.Cmd.Wait()
}
