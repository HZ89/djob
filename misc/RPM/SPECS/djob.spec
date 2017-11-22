%define srcname djob

Name: %{srcname}
Version: ##VERSION
Release: 1%{?dist}
Summary: djob is a distributed job scheduling system
License: GPLV3
URL: https://github.com/hz89/djob
Source0: https://github.com/HZ89/djob/releases/download/%{VERSION}/djob_linux_amd64.tar.gz

%description
djob supports cron syntax (second-level supported) for scheduled and one-time job. Select the
server to execute the job through the user-defined tags. Job is randomly assigned to the server in the 
same region for scheduling. If any one server failed, the job it managed will automatically take over 
by another.

%prep
%setup -c djob_linux_amd64.tar.gz


%build

%install
rm -rf %{buildroot}
%{__install} -p -D -m 0755 djob %{buildroot}%{_bindir}/djob
mkdir -p %{buildroot}%{_sysconfdir}/djob
cat <<EOF > %{buildroot}%{_sysconfdir}/djob/djob.yml
server: true   # start up as a server
load_job_policy: 'nothing'  # load policy, nothing, own, all
node: 'server1' # node name
region: 'test'  # node region
log_level: 'debug' # log level
sql_host: '' # sql store ip just support TiDB, MySQL
sql_port: 4000 # sql port
sql_user: '' # sql user
sql_dbname: 'djob' # database name
sql_password: ''
rpc_tls: true   # grpc use https
rpc_key_file: ''
rpc_cert_file: ''
rpc_ca_file: ''
kv_store: ''   # kv store, etcd, zk etc...
kv_store_servers: # kv store ip:port list
  - ''
kv_store_keyspeace: 'djob'
join:  # serf jion server list
  - ''
serf_bind_addr: ''
http_api_addr: ''
rpc_bind_addr: ''
tokens:   # api auth token
  defualt: 'VSTaJodQaxNdWC/ePjmY8Q=='
  test: '1sFVN7/zISAQWmYGTvQX0w=='
encrypt_key: '1sFVN7/zISAQWmYGTvQX0w=='
tags:   # tags, just support string
  zone: 'ns1'
  role: 'web'
  scrip_version: '1.0'
EOF

%post
%systemd_post %{name}.service

%postun
%systemd_postun_with_restart %{name}.service

%files
%defattr(0644, root, root, 0755)
%{_bindir}/djob
%config(noreplace) %{_sysconfdir}/djob/djob.yml

%changelog
* Wed Nov 22 2017 Harrion Zhu <wcg6121@gmail.com> - 0.2.1-1
- create this file