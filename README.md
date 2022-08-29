# GoCache

GoCache 使用 LRU 缓存淘汰算法，并使用 protobuf 库优化节点间通信的性能。

## 已实现

- LRU 缓存淘汰策略
- 一致性哈希算法选择远程节点
- 使用锁加哈希解决缓存击穿问题
- 添加 protobuf 优化节点间通信


