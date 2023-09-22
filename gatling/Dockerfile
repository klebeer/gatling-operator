FROM amazoncorretto:11

# working directory for gatling
WORKDIR /opt

# gating version
ENV GATLING_VERSION 3.2.1
ENV PATH $PATH:/opt/gradle/bin:/opt/gatling/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin
ENV GATLING_HOME /opt/gatling

# create directory for gatling install
RUN yum install -y tzdata && mkdir -p gatling

# install gatling
RUN yum install -y wget unzip && \
    mkdir -p /tmp/downloads && \
    wget -q -O /tmp/downloads/gatling-$GATLING_VERSION.zip \
    https://repo1.maven.org/maven2/io/gatling/highcharts/gatling-charts-highcharts-bundle/$GATLING_VERSION/gatling-charts-highcharts-bundle-$GATLING_VERSION-bundle.zip && \
    mkdir -p /tmp/archive && cd /tmp/archive && \
    unzip /tmp/downloads/gatling-$GATLING_VERSION.zip && \
    mv /tmp/archive/gatling-charts-highcharts-bundle-$GATLING_VERSION/* /opt/gatling/ && \
    rm -rf /opt/gatling/user-files/simulations/computerdatabase /tmp/*

# Install Gradle 8.2
RUN wget -q -O /tmp/gradle-8.2-bin.zip https://services.gradle.org/distributions/gradle-8.2-bin.zip && \
    unzip /tmp/gradle-8.2-bin.zip -d /opt && \
    mv /opt/gradle-8.2 /opt/gradle && \
    rm /tmp/gradle-8.2-bin.zip


# change context to gatling directory
WORKDIR  /opt/gatling

# set directories below to be mountable from host
VOLUME ["/opt/gatling/results"]

# copy local files to gatling directory
COPY gatling-gradle gatling-gradle/

RUN cd /opt/gatling/gatling-gradle && gradle dependencies --refresh-dependencies && gradle gatlingClasses



ENTRYPOINT ["gatling.sh"]

