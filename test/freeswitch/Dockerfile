# Credit: https://github.com/kovalyshyn/docker-freeswitch/blob/vanilla/Dockerfile

FROM webitel/freeswitch-base

RUN apt-get -y --quiet update \
    # Substantially stripped down package list (compared to freeswitch-all), to limit image size.
    && apt-get -y --quiet --force-yes install freeswitch freeswitch-conf-vanilla freeswitch-init freeswitch-lang freeswitch-mod-commands freeswitch-mod-console freeswitch-mod-sofia freeswitch-sounds freeswitch-timezones \
    && cp -a /usr/share/freeswitch/conf/vanilla /etc/freeswitch \
    && rm -rf /usr/share/freeswitch/conf \
    && mkdir -p /docker-entrypoint.d /certs /db /recordings /scripts /var/lib/freeswitch /var/run/freeswitch \
    && apt-get clean && rm -rf /var/lib/apt/lists/*

ENV LANG en_US.utf8

COPY docker-entrypoint.sh /

ENTRYPOINT ["/docker-entrypoint.sh"]

CMD ["freeswitch"]
