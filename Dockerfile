# Uses build context from GoReleaser - images are prebuilt
FROM scratch
COPY castkeeper /
COPY castkeeper.yml.example /etc/castkeeper/castkeeper.yml
EXPOSE 8080
CMD ["/castkeeper"]
