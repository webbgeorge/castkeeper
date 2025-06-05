# Uses build context from GoReleaser - binaries are prebuilt
FROM ubuntu:24.04
COPY castkeeper /castkeeper
COPY castkeeper.yml.example /etc/castkeeper/castkeeper.yml.example
COPY LICENSE /etc/castkeeper/LICENSE
EXPOSE 8080
ENTRYPOINT [ "/castkeeper" ]
CMD [ "serve"]
