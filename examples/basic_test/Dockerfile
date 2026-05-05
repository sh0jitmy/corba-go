FROM ubuntu:18.04
ENV DEBIAN_FRONTEND=noninteractive

RUN apt-get update && apt-get install -y \
    python \
    python-omniorb \
    omniorb-nameserver \
    omniidl \
    omniidl-python \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app
COPY test.idl .
COPY test_client.py .

RUN omniidl -bpython test.idl

ENTRYPOINT ["python", "test_client.py"]
