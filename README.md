# Gard
Gard is a high-performance object-retrieval, compression/decompression and caching solution for video assets for fast retrieval and low-energy consumption runtime. Consist of a cluster of microservices including a Go API that interfaces with a video codec engine, instance of Apache Ozone (storage), Apache Ignite (caching), and Apache Spark (ML video-processing, flagging and insights).
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

IN PROGRESS

