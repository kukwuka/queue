# queue

## Запуск 

```shell
  go run ./cmd -port=8081 -timeout=10s -queueMaxSize=3 -queuesMaxCount=3
```

или

```shell
  make up
```

использовал 2 библиотеки 
- github.com/google/uuid (для генерации строк)
- github.com/stretchr/testify (для тестирования)
- github.com/vektra/mockery/v2 (для генерации моков, руками писать уж больно долго)