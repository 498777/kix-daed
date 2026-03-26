# SPDX-License-Identifier: GPL-2.0-only
#
# Copyright (C) 2023 ImmortalWrt.org

include $(TOPDIR)/rules.mk

PKG_NAME:=daed
PKG_VERSION:=2026.03.26
DAED_VERSION:=daed-819e043
WING_VERSION:=wing-6bb6a31
CORE_VERSION:=core-b27fbd7
WING_HASH_SHORT:=$(shell echo $(WING_VERSION) | cut -d- -f2)
CORE_HASH_SHORT:=$(shell echo $(CORE_VERSION) | cut -d- -f2)
PKG_RELEASE:=1

PKG_SOURCE:=$(PKG_NAME)-$(PKG_VERSION).tar.gz
PKG_SOURCE_PROTO:=git
PKG_SOURCE_VERSION:=819e043a7972ef53477e1180a5310faf069e7871
PKG_SOURCE_URL:=https://github.com/daeuniverse/daed.git
PKG_MIRROR_HASH:=skip

PKG_LICENSE:=AGPL-3.0-only MIT
PKG_LICENSE_FILES:=LICENSE wing/LICENSE
PKG_MAINTAINER:=Tianling Shen <cnsztl@immortalwrt.org>

PKG_BUILD_DIR:=$(BUILD_DIR)/$(PKG_NAME)-$(PKG_VERSION)
PKG_BUILD_DEPENDS:=golang/host bpf-headers
PKG_BUILD_PARALLEL:=1
PKG_BUILD_FLAGS:=no-mips16

GO_PKG:=github.com/daeuniverse/dae-wing
GO_PKG_SOURCE_PATH:=$(PKG_BUILD_DIR)/wing
GO_PKG_BUILD_PKG:=github.com/daeuniverse/dae-wing

GO_PKG_LDFLAGS:= \
	-s -w -X '$(GO_PKG)/db.AppDescription=$(PKG_NAME) is a integration solution of dae, API and UI.'
GO_PKG_LDFLAGS_X:= \
	$(GO_PKG)/db.AppName=$(PKG_NAME) \
	$(GO_PKG)/db.AppVersion=$(DAED_VERSION)_$(WING_VERSION)_$(CORE_VERSION)
GO_PKG_TAGS:=embedallowed,trace

include $(INCLUDE_DIR)/package.mk
include $(INCLUDE_DIR)/bpf.mk
include $(TOPDIR)/feeds/packages/lang/golang/golang-package.mk

GO_PKG_TARGET_VARS+= \
	CGO_LDFLAGS="$(TARGET_LDFLAGS) -static" \
	GOEXPERIMENT=newinliner,simd
GO_PKG_DEFAULT_LDFLAGS:=-w -s -linkmode external -extldflags '-static'

define Package/daed/Default
  SECTION:=net
  CATEGORY:=Network
  SUBMENU:=Web Servers/Proxies
  URL:=https://github.com/daeuniverse/daed
endef

define Package/daed
  $(call Package/daed/Default)
  TITLE:=A Modern Dashboard For dae
  DEPENDS:=$(GO_ARCH_DEPENDS) \
    +ca-bundle +kmod-sched-core +kmod-sched-bpf +kmod-xdp-sockets-diag \
    +kmod-veth +v2ray-geoip +v2ray-geosite \
    +@KERNEL_XDP_SOCKETS \
    +DAED_USE_VMLINUX_BTF:vmlinux-btf
endef

define Package/daed/config
	choice
		prompt "BTF source for CO-RE"
		default DAED_USE_KERNEL_BTF
		depends on PACKAGE_daed

	config DAED_USE_KERNEL_BTF
		bool "Use kernel BTF (integrated)"
		depends on KERNEL_DEBUG_INFO_BTF

	config DAED_USE_VMLINUX_BTF
		bool "Use vmlinux-btf package"
	endchoice
endef

define Package/daed/description
  daed is a backend of dae, provides a method to bundle arbitrary
  frontend, dae and geodata into one binary.
endef

define Package/daed/conffiles
/etc/daed/wing.db
/etc/config/daed
endef

NODE_VERSION:=v24.12.0
NODE_URL:=https://nodejs.org/dist/$(NODE_VERSION)/node-$(NODE_VERSION)-linux-x64.tar.xz

define Build/Prepare
	$(TAR) --strip-components=1 -C $(PKG_BUILD_DIR) -xzf $(DL_DIR)/$(PKG_SOURCE)
	
	rm -rf $(PKG_BUILD_DIR)/wing && git clone https://github.com/daeuniverse/dae-wing $(PKG_BUILD_DIR)/wing
	cd $(PKG_BUILD_DIR)/wing && git checkout $(WING_HASH_SHORT)
	
	rm -rf $(PKG_BUILD_DIR)/wing/dae-core && git clone https://github.com/olicesx/dae $(PKG_BUILD_DIR)/wing/dae-core
	cd $(PKG_BUILD_DIR)/wing/dae-core && git checkout $(CORE_HASH_SHORT) && git submodule update --init && go mod tidy
	
	rm -rf $(PKG_BUILD_DIR)/outbound && git clone --depth=1 -b fix/direct-memory-safety https://github.com/olicesx/outbound.git $(PKG_BUILD_DIR)/outbound
	rm -rf $(PKG_BUILD_DIR)/quic-go && git clone --depth=1 -b main https://github.com/olicesx/quic-go.git $(PKG_BUILD_DIR)/quic-go
	
	( \
		cd $(PKG_BUILD_DIR)/wing ; \
		go mod edit -replace github.com/daeuniverse/outbound=../outbound ; \
		go mod edit -replace github.com/daeuniverse/quic-go=../quic-go ; \
		go mod tidy ; \
		go mod download ; \
	)

	( \
		mkdir -p $(PKG_BUILD_DIR)/.node_tmp ; \
		wget -qO - "$(NODE_URL)" | tar -xJ -C $(PKG_BUILD_DIR)/.node_tmp --strip-components=1 ; \
		export PATH="$(PKG_BUILD_DIR)/.node_tmp/bin:$$$$PATH" ; \
		npm install -g pnpm ; \
		cd $(PKG_BUILD_DIR) ; \
		pnpm install ; \
		pnpm build --filter daed ; \
		mkdir -p wing/webrender/web ; \
		cp -rf apps/web/dist/* wing/webrender/web/ ; \
		find wing/webrender/web -type f -size +4k ! -name "*.gz" ! -name "*.woff" ! -name "*.woff2" -exec sh -c '\
			gzip -9 -k "{}"; \
			if [ "$$$$(stat -c %s "{}")" -lt "$$$$(stat -c %s "{}.gz")" ]; then \
				rm "{}.gz"; \
			else \
				rm "{}"; \
			fi' ";" ; \
	)
	rm -rf $(PKG_BUILD_DIR)/.node_tmp
endef

DAE_CFLAGS:= \
	-O2 -Wall -Werror \
	-DMAX_MATCH_SET_LEN=1024 \
	-I$(BPF_HEADERS_DIR)/tools/lib \
	-I$(BPF_HEADERS_DIR)/arch/$(BPF_KARCH)/include/asm/mach-generic

define Build/Compile
	( \
		cd $(PKG_BUILD_DIR)/wing ; \
		export $(GO_PKG_VARS) ; \
		go generate ./... ; \
		cd dae-core ; \
		export \
			BPF_CLANG="$(CLANG)" \
			BPF_STRIP_FLAG="-strip=$(LLVM_STRIP)" \
			BPF_CFLAGS="$(DAE_CFLAGS)" \
			BPF_TARGET="bpfel,bpfeb" \
			BPF_TRACE_TARGET="$(GO_ARCH)" ; \
		go generate control/control.go ; \
		go generate trace/trace.go ; \
	)
	$(call GoPackage/Build/Compile)
endef

define Package/daed/install
	$(call GoPackage/Package/Install/Bin,$(PKG_INSTALL_DIR))
	$(INSTALL_DIR) $(1)/usr/bin
	$(INSTALL_BIN) $(PKG_INSTALL_DIR)/usr/bin/dae-wing $(1)/usr/bin/daed

	$(INSTALL_DIR) $(1)/etc/config
	$(INSTALL_CONF) $(CURDIR)/files/daed.config $(1)/etc/config/daed

	$(INSTALL_DIR) $(1)/etc/init.d
	$(INSTALL_BIN) $(CURDIR)/files/daed.init $(1)/etc/init.d/daed
endef

$(eval $(call GoBinPackage,daed))
$(eval $(call BuildPackage,daed))
