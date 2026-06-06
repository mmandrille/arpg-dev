MAKEDIR := $(dir $(abspath $(lastword $(MAKEFILE_LIST))))

include $(MAKEDIR)db.mk
include $(MAKEDIR)tools.mk
include $(MAKEDIR)server.mk
include $(MAKEDIR)shared.mk
include $(MAKEDIR)test.mk
include $(MAKEDIR)agents.mk
include $(MAKEDIR)client.mk
include $(MAKEDIR)ci.mk
