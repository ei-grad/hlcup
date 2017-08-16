FROM busybox
CMD /hlcup
EXPOSE 80
ENV RUN_TOP 1
ADD hlcup /
