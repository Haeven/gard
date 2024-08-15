# Gard
Gard is a high-performance object-retrieval, compression/decompression and caching solution for video assets for fast retrieval and low-energy consumption runtime. Consist of a cluster of microservices including a Rust API that interfaces with an container instance of Apache Ozone, a compression encoding and decoding service (respectively), caching volume/service. 

## How to run
Run following command in this folder as active directory (locate /README.md)

```

```

### Overview of Gard infrastructure
Gard core consist of the following key components:

- Storage Backend: Responsible for storing video files and vector objects.
- API Layer: RESTful APIs for interaction with the storage system.
- Metadata Management: Handle metadata associated with each stored object.
- Authentication & Authorization: Secure access to the system.
- Encryption & Integrity: Ensure data security and integrity.
- Concurrency & Scalability: Efficiently handle multiple requests.
- Indexing & Searching: Facilitate quick retrieval of stored objects.

### Architectural Overview

Here's how the architectural components fit together:

- Client Interaction: Clients interact with the API layer using HTTP/HTTPS requests.
- Storage Engine: Handles reading and writing of objects to disk.
- Metadata Database: Stores metadata such as file names, sizes, types, and tags.
- Security Layer: Manages authentication, authorization, encryption, and data integrity.
- Key Technologies
- Programming Language: Rust, for performance and safety.
- Storage Backend: Filesystem, or use a distributed storage solution like MinIO.
- Database: PostgreSQL or SQLite for metadata.
- HTTP Server: Actix Web or Warp for building RESTful APIs.
- Authentication: JWT tokens for secure access control.
- Encryption: AES encryption for data security.