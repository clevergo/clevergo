BENCHMARK
---------

- Date: 2020/03/13
- CPU: 4 Core
- RAM: 8G 
- Go: 1.14
- [Source](https://github.com/razonyang/go-http-routing-benchmark)

**Lower is better!**

[![Benchmark](https://razonyang.com/wp-content/uploads/2020/03/benchmark.png)](BENCHMARK.md)

```shell
BenchmarkEcho_Param                     14462064                87.1 ns/op             0 B/op          0 allocs/op
BenchmarkGin_Param                      13330479                94.1 ns/op             0 B/op          0 allocs/op
BenchmarkHttpRouter_Param                9816159               126 ns/op              32 B/op          1 allocs/op
BenchmarkCleverGo_Param                 15438733                78.3 ns/op             0 B/op          0 allocs/op
BenchmarkEcho_Param5                     5895441               232 ns/op               0 B/op          0 allocs/op
BenchmarkGin_Param5                      6737994               195 ns/op               0 B/op          0 allocs/op
BenchmarkHttpRouter_Param5               2849174               418 ns/op             160 B/op          1 allocs/op
BenchmarkCleverGo_Param5                 9524612               145 ns/op               0 B/op          0 allocs/op
BenchmarkEcho_Param20                    1681609               709 ns/op               0 B/op          0 allocs/op
BenchmarkGin_Param20                     2273236               522 ns/op               0 B/op          0 allocs/op
BenchmarkHttpRouter_Param20               968998              1392 ns/op             640 B/op          1 allocs/op
BenchmarkCleverGo_Param20                3698936               321 ns/op               0 B/op          0 allocs/op
BenchmarkEcho_ParamWrite                 6235656               215 ns/op               8 B/op          1 allocs/op
BenchmarkGin_ParamWrite                  7236438               186 ns/op               0 B/op          0 allocs/op
BenchmarkHttpRouter_ParamWrite           6906879               173 ns/op              32 B/op          1 allocs/op
BenchmarkCleverGo_ParamWrite            10565263               141 ns/op               0 B/op          0 allocs/op
BenchmarkEcho_GithubStatic               9273700               132 ns/op               0 B/op          0 allocs/op
BenchmarkGin_GithubStatic                7864270               136 ns/op               0 B/op          0 allocs/op
BenchmarkHttpRouter_GithubStatic        24161659                54.5 ns/op             0 B/op          0 allocs/op
BenchmarkCleverGo_GithubStatic          15532950                74.5 ns/op             0 B/op          0 allocs/op
BenchmarkEcho_GithubParam                5102701               261 ns/op               0 B/op          0 allocs/op
BenchmarkGin_GithubParam                 4683662               252 ns/op               0 B/op          0 allocs/op
BenchmarkHttpRouter_GithubParam          3040068               409 ns/op              96 B/op          1 allocs/op
BenchmarkCleverGo_GithubParam            8008634               165 ns/op               0 B/op          0 allocs/op
BenchmarkEcho_GithubAll                    26676             48345 ns/op               0 B/op          0 allocs/op
BenchmarkGin_GithubAll                     26343             46739 ns/op               0 B/op          0 allocs/op
BenchmarkHttpRouter_GithubAll              18895             64755 ns/op           13792 B/op        167 allocs/op
BenchmarkCleverGo_GithubAll                37891             30822 ns/op               0 B/op          0 allocs/op
BenchmarkEcho_GPlusStatic               14978847                78.1 ns/op             0 B/op          0 allocs/op
BenchmarkGin_GPlusStatic                12845145                87.9 ns/op             0 B/op          0 allocs/op
BenchmarkHttpRouter_GPlusStatic         38654924                30.1 ns/op             0 B/op          0 allocs/op
BenchmarkCleverGo_GPlusStatic           23368812                51.8 ns/op             0 B/op          0 allocs/op
BenchmarkEcho_GPlusParam                10425043               119 ns/op               0 B/op          0 allocs/op
BenchmarkGin_GPlusParam                  8474382               141 ns/op               0 B/op          0 allocs/op
BenchmarkHttpRouter_GPlusParam           5565493               230 ns/op              64 B/op          1 allocs/op
BenchmarkCleverGo_GPlusParam            11719412               111 ns/op               0 B/op          0 allocs/op
BenchmarkEcho_GPlus2Params               6037066               203 ns/op               0 B/op          0 allocs/op
BenchmarkGin_GPlus2Params                5424814               213 ns/op               0 B/op          0 allocs/op
BenchmarkHttpRouter_GPlus2Params         4450670               281 ns/op              64 B/op          1 allocs/op
BenchmarkCleverGo_GPlus2Params           8712440               136 ns/op               0 B/op          0 allocs/op
BenchmarkEcho_GPlusAll                    662982              2019 ns/op               0 B/op          0 allocs/op
BenchmarkGin_GPlusAll                     598694              2076 ns/op               0 B/op          0 allocs/op
BenchmarkHttpRouter_GPlusAll              530766              2836 ns/op             640 B/op         11 allocs/op
BenchmarkCleverGo_GPlusAll                951831              1299 ns/op               0 B/op          0 allocs/op
BenchmarkEcho_ParseStatic               14958830                81.0 ns/op             0 B/op          0 allocs/op
BenchmarkGin_ParseStatic                11871231                92.5 ns/op             0 B/op          0 allocs/op
BenchmarkHttpRouter_ParseStatic         39344148                30.4 ns/op             0 B/op          0 allocs/op
BenchmarkCleverGo_ParseStatic           23199598                60.2 ns/op             0 B/op          0 allocs/op
BenchmarkEcho_ParseParam                10894990                99.7 ns/op             0 B/op          0 allocs/op
BenchmarkGin_ParseParam                 10977565               113 ns/op               0 B/op          0 allocs/op
BenchmarkHttpRouter_ParseParam           5711146               219 ns/op              64 B/op          1 allocs/op
BenchmarkCleverGo_ParseParam            13230211                88.9 ns/op             0 B/op          0 allocs/op
BenchmarkEcho_Parse2Params               7259296               147 ns/op               0 B/op          0 allocs/op
BenchmarkGin_Parse2Params                8326382               142 ns/op               0 B/op          0 allocs/op
BenchmarkHttpRouter_Parse2Params         5601914               237 ns/op              64 B/op          1 allocs/op
BenchmarkCleverGo_Parse2Params          12411094               103 ns/op               0 B/op          0 allocs/op
BenchmarkEcho_ParseAll                    368058              3302 ns/op               0 B/op          0 allocs/op
BenchmarkGin_ParseAll                     332031              4503 ns/op               0 B/op          0 allocs/op
BenchmarkHttpRouter_ParseAll              354960              3831 ns/op             640 B/op         16 allocs/op
BenchmarkCleverGo_ParseAll                537031              2237 ns/op               0 B/op          0 allocs/op
BenchmarkEcho_StaticAll                    44497             29439 ns/op               0 B/op          0 allocs/op
BenchmarkGin_StaticAll                     42300             31039 ns/op               0 B/op          0 allocs/op
BenchmarkHttpRouter_StaticAll              95260             13101 ns/op               0 B/op          0 allocs/op
BenchmarkCleverGo_StaticAll                70975             16985 ns/op               0 B/op          0 allocs/op
PASS
ok      github.com/julienschmidt/go-http-routing-benchmark      94.614s
```