local key = KEYS[1]
local cntKey = key .. ":cnt"
local inputCode = ARGV[1]

local code = redis.call("get", key)
local cnt = tonumber(redis.call("get", cntKey))

-- 未发送、已过期，或者次数记录异常
if not code or not cnt then
    return -2
end

-- 已使用，或者三次错误机会已经耗尽
if cnt <= 0 then
    return -1
end

if code == inputCode then
    -- 保留验证码 key，避免绕过发送间隔。
    -- 次数标记失效，并跟随验证码一起过期。
    local ttl = redis.call("pttl", key)
    if ttl <= 0 then
        return -2
    end

    redis.call("set", cntKey, -1, "PX", ttl)
    return 0
end

redis.call("decr", cntKey)
return -2