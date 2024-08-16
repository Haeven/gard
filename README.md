# Gard
Gard is a high-performance object-retrieval, compression/decompression and caching solution for video assets for fast retrieval and low-energy consumption runtime. Consist of a cluster of microservices including a Go API that interfaces with a Rust video encoding and decoding engine (AV1), instance of Apache Ozone and Spark, an AV1 compression encoding and decoding service (respectively), caching volume/service (Dragonfly [Redis alt]). 

## How to run
To run the engine
```
cd engine
```
```
docker build -t gard-encoder .
```
```
docker run -p 8081:8081 gard-encoder
```

TODO 
