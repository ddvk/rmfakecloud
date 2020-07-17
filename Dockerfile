FROM scratch
EXPOSE 3000
#ENV 
COPY dist/rmfake-docker .
ENTRYPOINT ["/rmfake-docker"]
