FROM scratch
EXPOSE 3000
#ENV 
COPY bin/rmfake-docker .
ENTRYPOINT ["/rmfake-docker"]
