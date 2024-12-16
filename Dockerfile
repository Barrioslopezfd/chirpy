FROM debian:trixie-slim
ADD chirpy /bin/chirpy

CMD ["/bin/chirpy"]
