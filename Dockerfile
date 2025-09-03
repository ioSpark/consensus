FROM scratch

EXPOSE 8088
WORKDIR /app

COPY /consensus ./

ENTRYPOINT [ "./consensus" ]
