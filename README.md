Docker external graph driver
============================

# Why external graph driver
Long times ago, @dachary proposed implement [Ceph storage driver](https://github.com/docker/docker/issues/8854), but still not achieved now. The main reason is that it is not possible to statically compile docker with the ceph graph driver enabled because some static libraries are missing (Ubuntu 14.04) at the moment. 

For support rbd storage driver, must use dynamic compile.This is not accepted by docker community.

- [#9146](https://github.com/docker/docker/pull/9146)
- [#14800](https://github.com/docker/docker/pull/14800/)

Now docker community plan to implement out-of-process graph driver [#13777](https://github.com/docker/docker/pull/13777). It is a good tradeoff between Docker and Ceph.

On the other hand, companies like EMC, NetApp and others will most likely not be sending pull requests to add their product specific graph driver to the Docker repository. That can be because of a multitude of reasons: they want to keep it closed source, they want to put some proprietary stuff in there, they'd want to be able to change it, update it separately from Docker releases and so on.

# Support graph drivers

- [Ceph rbd driver](https://github.com/hustcat/docker-graph-driver/driver/rbd/README.md)

# How to use

TODO:
