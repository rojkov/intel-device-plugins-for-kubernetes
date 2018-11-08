FROM fedora:28 as builder
RUN dnf update -y && \
    dnf install -y wget make gcc-c++ findutils golang-bin && \
    mkdir -p /usr/src/qat && \
    cd /usr/src/qat && \
    wget https://01.org/sites/default/files/downloads/intelr-quickassist-technology/qat1.7.l.4.2.0-00012.tar.gz && \
    tar xf *.tar.gz
RUN cd /usr/src/qat/quickassist/utilities/adf_ctl && \
    make KERNEL_SOURCE_DIR=/usr/src/qat/quickassist/qat && \
    cp -a adf_ctl /usr/bin/
ARG DIR=/root/go/src/github.com/intel/intel-device-plugins-for-kubernetes
WORKDIR $DIR
COPY . .
RUN cd cmd/qat2_plugin; go install
RUN chmod a+x /root/go/bin/qat2_plugin

from fedora:28
RUN dnf update -y && \
    dnf install -y libstdc++
COPY --from=builder /root/go/bin/qat2_plugin /usr/bin/intel_qat2_device_plugin
COPY --from=builder /usr/bin/adf_ctl /usr/bin/adf_ctl
CMD ["/usr/bin/intel_qat2_device_plugin"]
