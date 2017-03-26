FROM alpine:3.4

ADD emoji-pod-visualizer /
ADD templates /templates
ADD images /images

WORKDIR /

CMD /emoji-pod-visualizer
