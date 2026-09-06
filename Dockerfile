# 基础镜像
FROM ubuntu:20.04
# 把编译后的大包进来这个镜像，放到工作目录 /app 随便换
COPY webook /app/webook
WORKDIR /app

# 最佳
# CMD 是执行命令
ENTRYPOINT ["/app/webook"]

