# Go Metrics

Проект по треку «Сервер сбора метрик и алертинга».

## Профилирование (iter17)

Написан [benchmark](./tests/bench_api_test.go) для тестирования API версий 1/2/3 сервера. Замеряется сценарий обновления метрик на сервере с отправкой с соотвествующей версией клиента. С него снимаются профили `cpu` и `mem` под каждую версию API (например `make bench-v1`).

- API V1 - отправка метрик, где type, name, value находятся в path параметрах
- API V2 - json payload под одну метрику
- API V3 - json payload под batch метрик

Оптимизация производилась на основе API V2 т.к. он показал худшие результаты.

Оптимизации:

- Замена стандартной библиотеки `compress` на `klauspost/compress` (агент + сервер)
- Реализация `CompressRoundTripper` с переиспользованием энкодеров из `sync.Pool` (агент)

    Результат: -91% суммарного потребления памяти на аллокации `flate.NewWriter`, что уменьшило нагрузку на GC и дало -66% суммарно по всем узлам

    ```txt
    $ go tool pprof -top -diff_base=profiles/base.mem-v2.pprof bin/bench profiles/result.mem-v2.pprof
    File: bench
    Build ID: c03fad1e9c9f35a4fa66aeba25e3cfc2c2c26232
    Type: alloc_space
    Time: 2026-10-22 01:58:04 MSK
    Showing nodes accounting for -50431.01MB, 89.07% of 56619.84MB total
    Dropped 277 nodes (cum <= 283.10MB)
          flat  flat%   sum%        cum   cum%
    -42556.41MB 75.16% 75.16% -51818.98MB 91.52%  compress/flate.NewWriter (inline)
    -8994.57MB 15.89% 91.05% -8994.57MB 15.89%  compress/flate.(*compressor).initDeflate (inline)
    -1908.41MB  3.37% 94.42% -1908.41MB  3.37%  compress/flate.(*dictDecoder).init (inline)
     1368.90MB  2.42% 92.00%  1368.90MB  2.42%  github.com/klauspost/compress/flate.(*dictDecoder).init (inline)
     1230.53MB  2.17% 89.83%  1230.53MB  2.17%  github.com/klauspost/compress/flate.(*fastGen).Reset
     1043.73MB  1.84% 87.98%  1872.15MB  3.31%  github.com/klauspost/compress/flate.NewWriter (inline)
     -618.95MB  1.09% 89.08%  -618.95MB  1.09%  net/http.init.func16
      520.38MB  0.92% 88.16%   520.38MB  0.92%  github.com/klauspost/compress/flate.newFastEnc (inline)
     -459.24MB  0.81% 88.97% -2367.66MB  4.18%  compress/flate.NewReader
     -301.66MB  0.53% 89.50%  -301.66MB  0.53%  compress/flate.(*huffmanEncoder).generate
      239.89MB  0.42% 89.08%   828.42MB  1.46%  github.com/klauspost/compress/flate.(*compressor).init
     -137.61MB  0.24% 89.32%  -269.21MB  0.48%  compress/flate.newHuffmanBitWriter (inline)
      129.27MB  0.23% 89.09%  1498.67MB  2.65%  github.com/klauspost/compress/flate.NewReaderOpts
       33.02MB 0.058% 89.03%  1693.82MB  2.99%  github.com/klauspost/compress/gzip.NewReader (inline)
      -26.52MB 0.047% 89.08% -1871.54MB  3.31%  compress/gzip.NewReader (inline)
        3.50MB 0.0062% 89.08% -33464.75MB 59.10%  io.copyBuffer
           2MB 0.0035% 89.07% -17075.27MB 30.16%  github.com/go-chi/chi/v5/middleware.ClientIPFromRemoteAddr.func1
          -1MB 0.0018% 89.07% -32178.35MB 56.83%  github.com/nikimonax/go-metrics/internal/lib/httpextra.(*CompressRoundTripper).compress
           1MB 0.0018% 89.07%   537.34MB  0.95%  github.com/go-chi/chi/v5/middleware.(*Compressor).selectEncoder
        0.64MB 0.0011% 89.07% -9262.58MB 16.36%  compress/flate.(*compressor).init
       -0.50MB 0.00088% 89.07% -16903.06MB 29.85%  github.com/go-chi/chi/v5/middleware.(*Compressor).Handler-fm.(*Compressor).Handler.func1
        0.50MB 0.00088% 89.07%  -647.97MB  1.14%  net/http.(*Request).write
        0.50MB 0.00088% 89.07% -17265.22MB 30.49%  net/http.(*conn).serve
    ```

    ```txt
    $ go tool pprof -top -cum  -diff_base=profiles/base.cpu-v2.pprof bin/bench profiles/result.cpu-v2.pprof
    File: bench
    Build ID: c03fad1e9c9f35a4fa66aeba25e3cfc2c2c26232
    Type: cpu
    Time: 2026-10-22 01:57:51 MSK
    Duration: 17.55s, Total samples = 34.60s (197.15%)
    Showing nodes accounting for -23.01s, 66.50% of 34.60s total
    Dropped 840 nodes (cum <= 0.17s)
          flat  flat%   sum%        cum   cum%
        -0.03s 0.087% 0.087%    -16.41s 47.43%  runtime.systemstack
             0     0% 0.087%    -13.49s 38.99%  runtime.gcBgMarkWorker
             0     0% 0.087%    -12.82s 37.05%  runtime.gcBgMarkWorker.func2
        -0.26s  0.75%  0.84%    -12.81s 37.02%  runtime.gcDrain
             0     0%  0.84%     -6.56s 18.96%  runtime.gcDrainMarkWorkerDedicated (inline)
             0     0%  0.84%     -6.25s 18.06%  runtime.gcDrainMarkWorkerIdle (inline)
        -0.05s  0.14%  0.98%     -5.81s 16.79%  runtime.mallocgc
        -0.01s 0.029%  1.01%     -5.30s 15.32%  github.com/nikimonax/go-metrics/internal/impl/gateway.(*HTTPMetricV2Gateway).Send
             0     0%  1.01%     -5.30s 15.32%  github.com/nikimonax/go-metrics/internal/impl/gateway.(*HTTPMetricV2Gateway).SendBatch
             0     0%  1.01%     -5.30s 15.32%  github.com/nikimonax/go-metrics/tests.BenchmarkAPI.func1
             0     0%  1.01%     -5.30s 15.32%  testing.(*B).runN
             0     0%  1.01%     -5.29s 15.29%  testing.(*B).launch
         0.01s 0.029%  0.98%     -5.22s 15.09%  github.com/nikimonax/go-metrics/internal/impl/gateway.(*HTTPMetricV2Gateway).makeRequest
        -0.01s 0.029%  1.01%     -5.05s 14.60%  compress/gzip.(*Writer).Write
        -0.33s  0.95%  1.97%     -5.03s 14.54%  runtime.scanSpan
             0     0%  1.97%     -5.01s 14.48%  compress/flate.NewWriter (inline)
        -1.65s  4.77%  6.73%     -4.97s 14.36%  runtime.scanObject
             0     0%  6.73%     -4.87s 14.08%  net/http.(*Client).send
        -0.01s 0.029%  6.76%     -4.85s 14.02%  net/http.send
             0     0%  6.76%     -4.84s 13.99%  github.com/nikimonax/go-metrics/internal/lib/httpextra.(*CompressRoundTripper).RoundTrip
             0     0%  6.76%     -4.84s 13.99%  net/http.(*Client).Do (inline)
         0.01s 0.029%  6.73%     -4.84s 13.99%  net/http.(*Client).do
             0     0%  6.73%     -4.42s 12.77%  github.com/nikimonax/go-metrics/internal/lib/httpextra.(*CompressRoundTripper).compress
        -0.02s 0.058%  6.79%     -4.19s 12.11%  runtime.mallocgcLarge
        -2.77s  8.01% 14.80%     -3.95s 11.42%  runtime.tryDeferToSpanScan
             0     0% 14.80%     -3.74s 10.81%  net/http.(*conn).serve
        -0.11s  0.32% 15.12%     -3.62s 10.46%  runtime.newobject
             0     0% 15.12%     -3.55s 10.26%  io.Copy (inline)
        -0.02s 0.058% 15.17%     -3.55s 10.26%  io.copyBuffer
    ```
