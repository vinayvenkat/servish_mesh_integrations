FROM alpine:3.8
RUN apk --no-cache add ca-certificates
WORKDIR /bin/
COPY bin/mygrpclogger .
ENTRYPOINT [ "/bin/mygrpclogger" ]
CMD [ "8888" ]
EXPOSE 8888
