wget https://partner-images.canonical.com/core/focal/current/ubuntu-focal-core-cloudimg-amd64-root.tar.gz # 下载基本系统镜像到本地

docker build -t nas_server:1229 # 名不要变，tag随便
docker run -it -p 9000:9000 . # dockerfile下，端口不要变
