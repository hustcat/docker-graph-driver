Docker external ceph rbd graph driver
=====================================

# Why external rbd graph driver

Long times ago, @dachary proposed implement [Ceph storage driver](https://github.com/docker/docker/issues/8854), but still not achieved now. The main reason is that it is not possible to statically compile docker with the ceph graph driver enabled because some static libraries are missing (Ubuntu 14.04) at the moment. 

For support rbd storage driver, must use dynamic compile:

```bash
./hack/make.sh dynbinary
```

This is not accepted by docker community.

- [#9146](https://github.com/docker/docker/pull/9146)

- [#14800](https://github.com/docker/docker/pull/14800/)

Now docker community plan to implement out-of-process graph driver [#13777](https://github.com/docker/docker/pull/13777). It is a good tradeoff between docker and rbd.

# How to use

## install ceph cluster
TODO:

## run rbd graph driver

```bash
# docker-graph-driver -s rbd
...
```

## run docker daemon

## pull images

```bash
# docker pull centos:latest
Pulling repository centos
7322fbe74aa5: Download complete 
f1b10cd84249: Download complete 
c852f6d61e65: Download complete 
Status: Downloaded newer image for centos:latest
```

## list rbd image

```bash
# rbd list
docker_image_7322fbe74aa5632b33a400959867c8ac4290e9c5112877a7754be70cfe5d66e9
docker_image_base_image
docker_image_c852f6d61e65cddf1e8af1f6cd7db78543bfb83cdcd36845541cf6d9dfef20a0
docker_image_f1b10cd842498c23d206ee0cbeaa9de8d2ae09ff3c7af2723a9e337a6965d639
```
## run container

```bash
# docker run -it --rm centos:latest /bin/bash
[root@290238155b54 /]#
```

```bash
# rbd list
docker_image_290238155b547852916b732e38bc4494375e1ed2837272e2940dfccc62691f6c
docker_image_290238155b547852916b732e38bc4494375e1ed2837272e2940dfccc62691f6c-init
docker_image_7322fbe74aa5632b33a400959867c8ac4290e9c5112877a7754be70cfe5d66e9
docker_image_base_image
docker_image_c852f6d61e65cddf1e8af1f6cd7db78543bfb83cdcd36845541cf6d9dfef20a0
docker_image_f1b10cd842498c23d206ee0cbeaa9de8d2ae09ff3c7af2723a9e337a6965d639
```

