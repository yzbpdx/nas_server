`wget https://partner-images.canonical.com/core/focal/current/ubuntu-focal-core-cloudimg-amd64-root.tar.gz` # 下载基本系统镜像到本地

`docker build -t nas_server:1229 .` # 名不要变，tag随便
`docker run -it -p 9000:9000 nas_server:1229` # 端口不要变

# 已完成功能

1. 登录注册，redis、sql
2. 文件查看，跳转，返回
3. 文件下载，支持大文件
4. 文件上传，支持大文件
5. 新建文件夹
6. 重命名文件或文件夹

# 待完成

1. 文件夹下载
2. 文件或文件夹删除
3. 下载及上传界面
4. 文件夹上传
5. 上传下载取消、暂停、断点续传