# Generate devel rpm
%global with_devel 1
# Build project from bundled dependencies
%global with_bundled 1
# Build with debug info rpm
%global with_debug 1
# Run tests in check section
%global with_check 1
# Generate unit-test rpm
%global with_unit_test 1

%if 0%{?with_debug}
%global _dwz_low_mem_die_limit 0
%else
%global debug_package   %{nil}
%endif

%if ! 0%{?gobuild:1}
%define gobuild(o:) go build -ldflags "${LDFLAGS:-} -B 0x$(head -c20 /dev/urandom|od -An -tx1|tr -d ' \\n')" -a -v -x %{?**};
%endif

%global provider        github
%global provider_tld    com
%global project         honeytrap
%global repo            honeytrap
# https://github.com/honeytrap/honeytrap
%global provider_prefix %{provider}.%{provider_tld}/%{project}/%{repo}
%global import_path     %{provider_prefix}
%global commit          3b05793a2c40b2d21252e2d0f9567872744cbd91
%global shortcommit     %(c=%{commit}; echo ${c:0:7})

Name:           honeytrap
Version:        0
Release:        0.1.20182002git%{shortcommit}%{?dist}
Summary:        Advanced Honeypot framework
License:        AGPLv3
URL:            https://%{provider_prefix}
Source0:        https://%{provider_prefix}/archive/%{commit}/%{repo}-%{shortcommit}.tar.gz

# e.g. el6 has ppc64 arch without gcc-go, so EA tag is required
ExclusiveArch:  %{?go_arches:%{go_arches}}%{!?go_arches:%{ix86} x86_64 aarch64 %{arm}}
# If go_compiler is not set to 1, there is no virtual provide. Use golang instead.
BuildRequires:  %{?go_compiler:compiler(go-compiler)}%{!?go_compiler:golang}

%if ! 0%{?with_bundled}
# Remaining dependencies not included in main packages
BuildRequires: golang(github.com/mimoo/disco/libdisco)
BuildRequires: golang(github.com/dgraph-io/badger)
BuildRequires: golang(github.com/fatih/color)
BuildRequires: golang(github.com/google/gopacket/pcap)
BuildRequires: golang(github.com/songgao/water)
BuildRequires: golang(github.com/BurntSushi/toml)
BuildRequires: golang(github.com/mattn/go-isatty)
BuildRequires: golang(github.com/google/gopacket/layers)
BuildRequires: golang(github.com/honeytrap/golxc)
BuildRequires: golang(github.com/Shopify/sarama)
BuildRequires: golang(github.com/fuyufjh/splunk-hec-go)
BuildRequires: golang(golang.org/x/crypto/ssh)
BuildRequires: golang(github.com/glycerine/rbuf)
BuildRequires: golang(github.com/honeytrap/protocol)
BuildRequires: golang(github.com/op/go-logging)
BuildRequires: golang(github.com/google/gopacket)
BuildRequires: golang(golang.org/x/sync/syncmap)
BuildRequires: golang(github.com/golang/protobuf/proto)
BuildRequires: golang(gopkg.in/olivere/elastic.v5)
BuildRequires: golang(github.com/pkg/profile)
BuildRequires: golang(github.com/boltdb/bolt)
BuildRequires: golang(github.com/songgao/packets/ethernet)
BuildRequires: golang(golang.org/x/crypto/ssh/terminal)
BuildRequires: golang(github.com/elazarl/go-bindata-assetfs)
BuildRequires: golang(github.com/gorilla/websocket)
BuildRequires: golang(github.com/rs/xid)
BuildRequires: golang(github.com/dimfeld/httptreemux)
BuildRequires: golang(github.com/satori/go.uuid)
BuildRequires: golang(github.com/google/gopacket/pcapgo)
BuildRequires: golang(github.com/oschwald/maxminddb-golang)
%endif

%description
Honeytrap is an extensible and opensource system for running, monitoring and
managing honeypots.

%if 0%{?with_devel}
%package -n golang-%{provider}-%{project}-%{repo}-devel
Summary:       %{summary}
BuildArch:     noarch

%if 0%{?with_check} && ! 0%{?with_bundled}
BuildRequires: golang(github.com/BurntSushi/toml)
BuildRequires: golang(github.com/Shopify/sarama)
BuildRequires: golang(github.com/boltdb/bolt)
BuildRequires: golang(github.com/dgraph-io/badger)
BuildRequires: golang(github.com/dimfeld/httptreemux)
BuildRequires: golang(github.com/elazarl/go-bindata-assetfs)
BuildRequires: golang(github.com/fatih/color)
BuildRequires: golang(github.com/fuyufjh/splunk-hec-go)
BuildRequires: golang(github.com/glycerine/rbuf)
BuildRequires: golang(github.com/golang/protobuf/proto)
BuildRequires: golang(github.com/google/gopacket)
BuildRequires: golang(github.com/google/gopacket/layers)
BuildRequires: golang(github.com/google/gopacket/pcap)
BuildRequires: golang(github.com/google/gopacket/pcapgo)
BuildRequires: golang(github.com/gorilla/websocket)
BuildRequires: golang(github.com/honeytrap/golxc)
BuildRequires: golang(github.com/honeytrap/protocol)
BuildRequires: golang(github.com/mattn/go-isatty)
BuildRequires: golang(github.com/mimoo/disco/libdisco)
BuildRequires: golang(github.com/op/go-logging)
BuildRequires: golang(github.com/oschwald/maxminddb-golang)
BuildRequires: golang(github.com/pkg/profile)
BuildRequires: golang(github.com/rs/xid)
BuildRequires: golang(github.com/satori/go.uuid)
BuildRequires: golang(github.com/songgao/packets/ethernet)
BuildRequires: golang(github.com/songgao/water)
BuildRequires: golang(golang.org/x/crypto/ssh)
BuildRequires: golang(golang.org/x/crypto/ssh/terminal)
BuildRequires: golang(golang.org/x/sync/syncmap)
BuildRequires: golang(gopkg.in/olivere/elastic.v5)
%endif

Requires:      golang(github.com/BurntSushi/toml)
Requires:      golang(github.com/Shopify/sarama)
Requires:      golang(github.com/boltdb/bolt)
Requires:      golang(github.com/dgraph-io/badger)
Requires:      golang(github.com/dimfeld/httptreemux)
Requires:      golang(github.com/elazarl/go-bindata-assetfs)
Requires:      golang(github.com/fatih/color)
Requires:      golang(github.com/fuyufjh/splunk-hec-go)
Requires:      golang(github.com/glycerine/rbuf)
Requires:      golang(github.com/golang/protobuf/proto)
Requires:      golang(github.com/google/gopacket)
Requires:      golang(github.com/google/gopacket/layers)
Requires:      golang(github.com/google/gopacket/pcap)
Requires:      golang(github.com/google/gopacket/pcapgo)
Requires:      golang(github.com/gorilla/websocket)
Requires:      golang(github.com/honeytrap/golxc)
Requires:      golang(github.com/honeytrap/protocol)
Requires:      golang(github.com/mattn/go-isatty)
Requires:      golang(github.com/mimoo/disco/libdisco)
Requires:      golang(github.com/op/go-logging)
Requires:      golang(github.com/oschwald/maxminddb-golang)
Requires:      golang(github.com/pkg/profile)
Requires:      golang(github.com/rs/xid)
Requires:      golang(github.com/satori/go.uuid)
Requires:      golang(github.com/songgao/packets/ethernet)
Requires:      golang(github.com/songgao/water)
Requires:      golang(golang.org/x/crypto/ssh)
Requires:      golang(golang.org/x/crypto/ssh/terminal)
Requires:      golang(golang.org/x/sync/syncmap)
Requires:      golang(gopkg.in/olivere/elastic.v5)

Provides:      golang(%{import_path}/cmd) = %{version}-%{release}
Provides:      golang(%{import_path}/cmd/honeytrap) = %{version}-%{release}
Provides:      golang(%{import_path}/config) = %{version}-%{release}
Provides:      golang(%{import_path}/director) = %{version}-%{release}
Provides:      golang(%{import_path}/director/forward) = %{version}-%{release}
Provides:      golang(%{import_path}/director/lxc) = %{version}-%{release}
Provides:      golang(%{import_path}/director/qemu) = %{version}-%{release}
Provides:      golang(%{import_path}/event) = %{version}-%{release}
Provides:      golang(%{import_path}/listener) = %{version}-%{release}
Provides:      golang(%{import_path}/listener/agent) = %{version}-%{release}
Provides:      golang(%{import_path}/listener/canary) = %{version}-%{release}
Provides:      golang(%{import_path}/listener/canary/arp) = %{version}-%{release}
Provides:      golang(%{import_path}/listener/canary/ethernet) = %{version}-%{release}
Provides:      golang(%{import_path}/listener/canary/icmp) = %{version}-%{release}
Provides:      golang(%{import_path}/listener/canary/ipv4) = %{version}-%{release}
Provides:      golang(%{import_path}/listener/canary/tcp) = %{version}-%{release}
Provides:      golang(%{import_path}/listener/canary/udp) = %{version}-%{release}
Provides:      golang(%{import_path}/listener/netstack) = %{version}-%{release}
Provides:      golang(%{import_path}/listener/socket) = %{version}-%{release}
Provides:      golang(%{import_path}/listener/tap) = %{version}-%{release}
Provides:      golang(%{import_path}/listener/tun) = %{version}-%{release}
Provides:      golang(%{import_path}/protocol) = %{version}-%{release}
Provides:      golang(%{import_path}/pushers) = %{version}-%{release}
Provides:      golang(%{import_path}/pushers/console) = %{version}-%{release}
Provides:      golang(%{import_path}/pushers/elasticsearch) = %{version}-%{release}
Provides:      golang(%{import_path}/pushers/eventbus) = %{version}-%{release}
Provides:      golang(%{import_path}/pushers/file) = %{version}-%{release}
Provides:      golang(%{import_path}/pushers/kafka) = %{version}-%{release}
Provides:      golang(%{import_path}/pushers/raven) = %{version}-%{release}
Provides:      golang(%{import_path}/pushers/slack) = %{version}-%{release}
Provides:      golang(%{import_path}/pushers/splunk) = %{version}-%{release}
Provides:      golang(%{import_path}/server) = %{version}-%{release}
Provides:      golang(%{import_path}/server/profiler) = %{version}-%{release}
Provides:      golang(%{import_path}/services) = %{version}-%{release}
Provides:      golang(%{import_path}/services/decoder) = %{version}-%{release}
Provides:      golang(%{import_path}/services/elasticsearch) = %{version}-%{release}
Provides:      golang(%{import_path}/services/ethereum) = %{version}-%{release}
Provides:      golang(%{import_path}/services/ftp) = %{version}-%{release}
Provides:      golang(%{import_path}/services/ipp) = %{version}-%{release}
Provides:      golang(%{import_path}/services/redis) = %{version}-%{release}
Provides:      golang(%{import_path}/services/ssh) = %{version}-%{release}
Provides:      golang(%{import_path}/services/vnc) = %{version}-%{release}
Provides:      golang(%{import_path}/sniffer) = %{version}-%{release}
Provides:      golang(%{import_path}/storage) = %{version}-%{release}
Provides:      golang(%{import_path}/utils) = %{version}-%{release}
Provides:      golang(%{import_path}/utils/files) = %{version}-%{release}
Provides:      golang(%{import_path}/utils/tests) = %{version}-%{release}
Provides:      golang(%{import_path}/web) = %{version}-%{release}

%description -n golang-%{provider}-%{project}-%{repo}-devel
%{summary}

This package contains library source intended for
building other packages which use import path with
%{import_path} prefix.
%endif

%if 0%{?with_unit_test} && 0%{?with_devel}
%package -n golang-%{provider}-%{project}-%{repo}-unit-test-devel
Summary:         Unit tests for %{name} package

# test subpackage tests code from devel subpackage
Requires:        golang-%{provider}-%{project}-%{repo}-devel = %{version}-%{release}


%description -n golang-%{provider}-%{project}-%{repo}-unit-test-devel
%{summary}

This package contains unit tests for project
providing packages with %{import_path} prefix.
%endif

%prep
%setup -q -n %{repo}-%{commit}

%build
mkdir -p src/%{provider}.%{provider_tld}/%{project}
ln -s ../../../ src/%{import_path}

export GOPATH=$(pwd):%{gopath}

%gobuild -o bin/honeytrap %{import_path}/
#%%gobuild -o bin/gen-ldflags %%{import_path}/scripts

%install
install -d -p %{buildroot}%{_bindir}
install -p -m 0755 bin/honeytrap %{buildroot}%{_bindir}

# source codes for building projects
%if 0%{?with_devel}
install -d -p %{buildroot}/%{gopath}/src/%{import_path}/
echo "%%dir %%{gopath}/src/%%{import_path}/." >> devel.file-list
# find all *.go but no *_test.go files and generate devel.file-list
for file in $(find . \( -iname "*.go" -or -iname "*.s" \) \! -iname "*_test.go" | grep -v "vendor") ; do
    dirprefix=$(dirname $file)
    install -d -p %{buildroot}/%{gopath}/src/%{import_path}/$dirprefix
    cp -pav $file %{buildroot}/%{gopath}/src/%{import_path}/$file
    echo "%%{gopath}/src/%%{import_path}/$file" >> devel.file-list

    while [ "$dirprefix" != "." ]; do
        echo "%%dir %%{gopath}/src/%%{import_path}/$dirprefix" >> devel.file-list
        dirprefix=$(dirname $dirprefix)
    done
done
%endif

# testing files for this project
%if 0%{?with_unit_test} && 0%{?with_devel}
install -d -p %{buildroot}/%{gopath}/src/%{import_path}/
# find all *_test.go files and generate unit-test-devel.file-list
for file in $(find . -iname "*_test.go" | grep -v "vendor") ; do
    dirprefix=$(dirname $file)
    install -d -p %{buildroot}/%{gopath}/src/%{import_path}/$dirprefix
    cp -pav $file %{buildroot}/%{gopath}/src/%{import_path}/$file
    echo "%%{gopath}/src/%%{import_path}/$file" >> unit-test-devel.file-list

    while [ "$dirprefix" != "." ]; do
        echo "%%dir %%{gopath}/src/%%{import_path}/$dirprefix" >> devel.file-list
        dirprefix=$(dirname $dirprefix)
    done
done
%endif

%if 0%{?with_devel}
sort -u -o devel.file-list devel.file-list
%endif

%check
%if 0%{?with_check} && 0%{?with_unit_test} && 0%{?with_devel}
%if ! 0%{?with_bundled}
export GOPATH=%{buildroot}/%{gopath}:%{gopath}
%else
# Since we aren't packaging up the vendor directory we need to link
# back to it somehow. Hack it up so that we can add the vendor
# directory from BUILD dir as a gopath to be searched when executing
# tests from the BUILDROOT dir.
ln -s ./ ./vendor/src # ./vendor/src -> ./vendor

export GOPATH=%{buildroot}/%{gopath}:$(pwd)/vendor:%{gopath}
%endif

%if ! 0%{?gotest:1}
%global gotest go test
%endif

%gotest %{import_path}/pushers/console
%gotest %{import_path}/pushers/elasticsearch
%gotest %{import_path}/pushers/file
%gotest %{import_path}/pushers/kafka
%gotest %{import_path}/pushers/raven
%gotest %{import_path}/pushers/splunk
%gotest %{import_path}/services
%gotest %{import_path}/services/decoder
%gotest %{import_path}/services/ftp
%gotest %{import_path}/services/ipp
%gotest %{import_path}/services/vnc
%endif

#define license tag if not already defined
%{!?_licensedir:%global license %doc}

%files
%license LICENSE
%doc CONTRIBUTING.md README.md
%{_bindir}/honeytrap
#%%{_bindir}/gen-ldflags

%if 0%{?with_devel}
%files -n golang-%{provider}-%{project}-%{repo}-devel -f devel.file-list
%license LICENSE
%doc CONTRIBUTING.md README.md
%dir %{gopath}/src/%{provider}.%{provider_tld}/%{project}
%endif

%if 0%{?with_unit_test} && 0%{?with_devel}
%files -n golang-%{provider}-%{project}-%{repo}-unit-test-devel -f unit-test-devel.file-list
%license LICENSE
%doc CONTRIBUTING.md README.md
%endif

%changelog
* Tue Feb 20 2018 Athos Ribeiro <athoscr@fedoraproject.org> - 0-0.1.git3b05793
- Initial package with bundled dependencies

