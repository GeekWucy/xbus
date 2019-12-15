package services

import (
	"fmt"
	"regexp"

	"github.com/infrmods/xbus/utils"
)

var rValidName = regexp.MustCompile(`(?i)^[a-z][a-z0-9_.-]{5,}$`)
var rValidService = regexp.MustCompile(`(?i)^[a-z][a-z0-9_.-]{5,}:[a-z0-9][a-z0-9_.-]*$`)
var rValidZone = regexp.MustCompile(`(?i)^[a-z0-9][a-z0-9_-]{3,}$`)

func checkName(name string) error {
	if !rValidName.MatchString(name) {
		return utils.NewError(utils.EcodeInvalidName, "")
	}
	return nil
}

func checkService(service string) error {
	if !rValidService.MatchString(service) {
		return utils.NewError(utils.EcodeInvalidService, "")
	}
	return nil
}

func checkServiceZone(service, zone string) error {
	if !rValidService.MatchString(service) {
		return utils.NewError(utils.EcodeInvalidService, "")
	}
	if !rValidZone.MatchString(zone) {
		return utils.NewError(utils.EcodeInvalidZone, "")
	}
	return nil
}

var rValidAddress = regexp.MustCompile(`(?i)^[a-z0-9:_.-]+$`)

func (ctrl *ServiceCtrl) checkAddress(addr string) error {
	if addr == "" {
		return utils.NewError(utils.EcodeInvalidEndpoint, "missing address")
	}
	if !rValidAddress.MatchString(addr) {
		return utils.NewError(utils.EcodeInvalidAddress, "")
	}
	if ctrl.config.isAddressBanned(addr) {
		return utils.NewError(utils.EcodeInvalidAddress, "banned")
	}
	return nil
}

func (ctrl *ServiceCtrl) serviceEntryPrefix(name string) string {
	return fmt.Sprintf("%s/%s/", ctrl.config.KeyPrefix, name)
}

const serviceDescNodeKey = "desc"

func (ctrl *ServiceCtrl) serviceDescKey(service, zone string) string {
	return fmt.Sprintf("%s/%s/%s/desc", ctrl.config.KeyPrefix, service, zone)
}

const serviceKeyNodePrefix = "node_"

func (ctrl *ServiceCtrl) serviceKey(service, zone, addr string) string {
	return fmt.Sprintf("%s/%s/%s/node_%s", ctrl.config.KeyPrefix, service, zone, addr)
}
