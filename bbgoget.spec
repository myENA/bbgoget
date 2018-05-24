%define debug_package %{nil}

Name:           bbgoget
Version:        0.1.2
Release:        1%{?dist}
Summary:        This makes bitbucket server private repos work with go get
BuildRoot:      %{_tmppath}/%{name}-%{version}-build


Group:          System Environment/Daemons
License:        MIT

%define         pkgpath github.com/myENA/%{name}
URL:            https://github.com/myENA/bbgoget

%undefine _disable_source_fetch
Source0:        https://%{pkgpath}/archive/v%{version}.tar.gz
%define         SHA256SUM0 97ff241f2052c5e0de6f20674a7b088504525974ae0f09898df53676427779cf
BuildRequires:  systemd-units golang

%define         bbgoget_user bbgoget
%define         bbgoget_group bbgoget
%define         bbgoget_home /etc/sysconfig

Requires(pre):      shadow-utils
Requires(post):     systemd
Requires(preun):    systemd
Requires(postun):   systemd

%description
bitbucket server private repos + go get = gravy

%pre
getent group %{bbgoget_group} >/dev/null || \
    groupadd -r %{bbgoget_group}
getent passwd %{bbgoget_user} >/dev/null || \
    useradd -r -g %{bbgoget_user} -d %{bbgoget_home} \
    -s /sbin/nologin -c %{name} %{bbgoget_user}
exit 0


%prep
echo "%{SHA256SUM0} %{SOURCE0}" | sha256sum -c -

%setup -q
mkdir -p go/{src,pkg,bin}
mkdir -p go/src/$(dirname %{pkgpath})
ln -s $(pwd) go/src/%{pkgpath}


%build
export GOPATH=$(pwd)/go
cd go/src/%{pkgpath}
pwd
ls -la
go build


%install
%{__install} -p -D -m 0644 go/src/%{pkgpath}/%{name}.service %{buildroot}%{_unitdir}/%{name}.service

%{__install} -p -D -m 0644 go/src/%{pkgpath}/%{name}.sysconfig %{buildroot}%{_sysconfdir}/sysconfig/%{name}

%{__install} -p -D -m 0755 go/src/%{pkgpath}/%{name} %{buildroot}%{_bindir}/%{name}


%post
%systemd_post %{name}.service

%preun
%systemd_preun %{name}.service

%postun
%systemd_postun_with_restart %{name}.service

%files
%defattr(-,root,root,-)
%{_unitdir}/%{name}.service
%{_bindir}/%{name}
%config(noreplace) %{_sysconfdir}/sysconfig/%{name}

%changelog
