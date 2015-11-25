FROM scratch
ADD main /
ADD conf.gcfg /
CMD ["/main"]
