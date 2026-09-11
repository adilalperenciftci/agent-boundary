// SPDX-License-Identifier: GPL-2.0-only OR BSD-2-Clause
#include "vmlinux.h"
#include <bpf/bpf_core_read.h>
#include <bpf/bpf_helpers.h>

#define TASK_COMM_LEN 16
#define PATH_LEN 256

struct exec_event {
    __u64 monotonic_ns;
    __u64 cgroup_id;
    __u64 start_time_ns;
    __u64 parent_start_time_ns;
    __u32 pid;
    __u32 tgid;
    __u32 ppid;
    __u32 uid;
    __u32 gid;
    __u32 pid_namespace;
    __u32 mount_namespace;
    __u32 parent_pid_namespace;
    char comm[TASK_COMM_LEN];
    char filename[PATH_LEN];
};

struct trace_event_raw_sched_process_exec___local {
    struct trace_entry ent;
    __u32 __data_loc_filename;
    pid_t pid;
    pid_t old_pid;
    char __data[0];
};

const volatile __u64 target_cgroup_id = 0;

struct {
    __uint(type, BPF_MAP_TYPE_RINGBUF);
    __uint(max_entries, 1 << 20);
} events SEC(".maps");

struct {
    __uint(type, BPF_MAP_TYPE_PERCPU_ARRAY);
    __uint(max_entries, 1);
    __type(key, __u32);
    __type(value, __u64);
} ringbuf_drops SEC(".maps");

static __always_inline __u32 active_pid_namespace(struct task_struct *task)
{
    struct pid *thread_pid = BPF_CORE_READ(task, thread_pid);
    if (!thread_pid)
        return 0;

    __u32 level = BPF_CORE_READ(thread_pid, level);
    if (level > 32)
        return 0;

    struct pid_namespace *namespace;
    bpf_core_read(&namespace, sizeof(namespace), &thread_pid->numbers[level].ns);
    if (!namespace)
        return 0;
    return BPF_CORE_READ(namespace, ns.inum);
}

static __always_inline __u32 mount_namespace(struct task_struct *task)
{
    struct nsproxy *proxy = BPF_CORE_READ(task, nsproxy);
    if (!proxy)
        return 0;
    struct mnt_namespace *namespace = BPF_CORE_READ(proxy, mnt_ns);
    if (!namespace)
        return 0;
    return BPF_CORE_READ(namespace, ns.inum);
}

SEC("tracepoint/sched/sched_process_exec")
int observe_exec(struct trace_event_raw_sched_process_exec___local *ctx)
{
    __u64 cgroup_id = bpf_get_current_cgroup_id();
    if (target_cgroup_id == 0 || cgroup_id != target_cgroup_id)
        return 0;

    struct exec_event *event = bpf_ringbuf_reserve(&events, sizeof(*event), 0);
    if (!event) {
        __u32 key = 0;
        __u64 *count = bpf_map_lookup_elem(&ringbuf_drops, &key);
        if (count)
            __sync_fetch_and_add(count, 1);
        return 0;
    }

    __builtin_memset(event, 0, sizeof(*event));
    __u64 pid_tgid = bpf_get_current_pid_tgid();
    __u64 uid_gid = bpf_get_current_uid_gid();
    struct task_struct *task = (struct task_struct *)bpf_get_current_task_btf();
    struct task_struct *parent = BPF_CORE_READ(task, real_parent);

    event->monotonic_ns = bpf_ktime_get_ns();
    event->cgroup_id = cgroup_id;
    event->start_time_ns = BPF_CORE_READ(task, start_boottime);
    event->parent_start_time_ns = BPF_CORE_READ(parent, start_boottime);
    event->pid = (__u32)pid_tgid;
    event->tgid = pid_tgid >> 32;
    event->ppid = BPF_CORE_READ(parent, tgid);
    event->uid = (__u32)uid_gid;
    event->gid = uid_gid >> 32;
    event->pid_namespace = active_pid_namespace(task);
    event->mount_namespace = mount_namespace(task);
    event->parent_pid_namespace = active_pid_namespace(parent);
    bpf_get_current_comm(event->comm, sizeof(event->comm));

    __u32 filename_offset = ctx->__data_loc_filename & 0xffff;
    bpf_probe_read_str(event->filename, sizeof(event->filename), (void *)ctx + filename_offset);
    bpf_ringbuf_submit(event, 0);
    return 0;
}

char LICENSE[] SEC("license") = "Dual BSD/GPL";
