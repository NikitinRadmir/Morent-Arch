# Целевая структура монорепозитория (план)

> Статус: **запланировано**, не применено. Переезд — отдельный PR `chore: move services under apps/`.

```text
Morent-Architecture/
├── README.md
├── Makefile
├── apps/
│   ├── morent/           # Morent-project
│   ├── user-system/      # user-system-develop
│   ├── car-aggregator/
│   ├── payment/
│   ├── email/
│   └── generator/
├── packages/
│   └── morent-events/    # shared/morent-events
├── deploy/
│   └── kafka/            # infra/kafka
└── docs/
```

Правило: `apps/` — runnable, `packages/` — библиотеки, `deploy/` — общая инфра.
