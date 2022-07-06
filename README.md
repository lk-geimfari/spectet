## Spectet

Host availability monitoring utility that supports various protocols.

Currently, `spectet` support the following protocols:

- `TCP`
- `UDP`
- `DNS`
- `HTTP`
- `ICMP` (root privileges required)

## How it works?

You have to create the JSON API endpoint with the following structure:

```json
[
  {
    "task_id": "UUID4",
    "hostname": "google.com",
    "port": 443,
    "task_type": "tcp"
  }
]
```

where `task_type` type should be one of: (`tcp`, `udp`, `resolve`, `http`, `icmp`).

After that you have to change `TasksApiURL` constant in `main.go`.

After that, you need to add the binary of `spectet` to cron with a schedule that fits you.
