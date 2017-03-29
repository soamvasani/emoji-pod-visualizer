FROM alpine:3.4

ADD emoji-pod-visualizer /
ADD templates /templates
ADD static /static

WORKDIR /

CMD /emoji-pod-visualizer
