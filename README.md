# Redis-GO
Making redis clone using GO

(progess status 1)
-> added SET\GET\TTL
-> created TCP server (listens on port:3679) | new GOroutine per client
-> implemented RESP (Redis Serialization Protocol) | improve command parsing and response formatting | pipelining is supported
