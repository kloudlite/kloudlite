{{ required ".kloudliteRelease must be set" .Values.kloudliteRelease }}
{{ required ".kloudliteDNSSuffix must be set" .Values.kloudliteDNSSuffix }}
{{ required ".certManager.clusterIssuer.acmeEmail must be set" .Values.certManager.clusterIssuer.acmeEmail }}

{{ required "when nginxIngress.wildcardCert.enabled is true, nginxIngress.wildcardCert.secretName must be set" (or .Values.nginxIngress.wildcardCert.enabled (and .Values.nginxIngress.wildcardCert.secretName .Values.nginxIngress.wildcardCert.secretNamespace)) }}
