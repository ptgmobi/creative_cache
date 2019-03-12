# creative_id
提供素材id查询服务，是一个缓存服务对外是`offer_update`来请求id，对内是跟`creative_center`交互

# 服务器
* ip: 52.221.198.147
* path: /opt/creative_cache
* 重启：cd /opt/creative_cache; bash start

# 服务逻辑
该服务本身主要是通过redis进行缓存，请求来了会去缓存里查，命中就直接返回，如果缓存没有就去`creative-center`里
查，如果有就保存如redis， 如果没有就放弃，而`creative-center`会在数据库里查询如果是新素材会去下载并加载信息
一方面以后使用。
