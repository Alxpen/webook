-- KEYS[1]: 当前 IP 的限流 key
-- ARGV[1]: 窗口长度（毫秒）；ARGV[2]: 最大请求数；ARGV[3]: 唯一请求标识
-- 返回 0 表示放行；正数表示距离可重试还需等待的毫秒数。
local key = KEYS[1]
local window = tonumber(ARGV[1])
local threshold = tonumber(ARGV[2])

-- 使用 Redis 时间，避免多个应用实例的时钟偏差。
local clock = redis.call('TIME')
local now = tonumber(clock[1]) * 1000 + math.floor(tonumber(clock[2]) / 1000)

-- 保留 (now - window, now] 内已放行的请求。
redis.call('ZREMRANGEBYSCORE', key, '-inf', now - window)
local count = redis.call('ZCARD', key)
if count >= threshold then
    local oldest = redis.call('ZRANGE', key, 0, 0, 'WITHSCORES')
    return math.max(1, tonumber(oldest[2]) + window - now)
end

redis.call('ZADD', key, now, ARGV[3])
-- 不再访问的 IP 自动清理；被拒绝的请求不延长窗口。
redis.call('PEXPIRE', key, window)
return 0
