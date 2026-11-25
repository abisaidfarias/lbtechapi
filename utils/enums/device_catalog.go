package enums

import "strings"

// DeviceCatalogCategory describes one device catalog: its label and group on
// the Catalogs screen, the device field that stores its values and the
// legacy configuration key it replaces.
type DeviceCatalogCategory struct {
	Key         string
	Label       string
	Group       string
	DeviceField string // bson field on devices
	ConfigField string // bson field on configurations (legacy source)
}

const (
	DeviceCatalogDeviceType       = "device_type"
	DeviceCatalogRom              = "rom"
	DeviceCatalogRam              = "ram"
	DeviceCatalogMemoryCapability = "memory_capability"
	DeviceCatalogUsb              = "usb"
)

// DeviceCatalogCategories is the fixed list, in screen order.
var DeviceCatalogCategories = []DeviceCatalogCategory{
	{DeviceCatalogDeviceType, "Device type", "General", "type", "device_types"},
	{DeviceCatalogRom, "ROM", "Memory", "memory_rom", "roms"},
	{DeviceCatalogRam, "RAM", "Memory", "memory_ram", "rams"},
	{DeviceCatalogMemoryCapability, "Memory capabilities", "Memory", "memory_cpu", "memory_capabilities"},
	{DeviceCatalogUsb, "USB", "Communications", "communication_usb", "usb_communications"},
}

// DeviceCatalogCategoryByKey returns the category definition.
func DeviceCatalogCategoryByKey(key string) (DeviceCatalogCategory, bool) {
	for _, c := range DeviceCatalogCategories {
		if c.Key == key {
			return c, true
		}
	}
	return DeviceCatalogCategory{}, false
}

// DeviceCatalogValueKey is the uniqueness key: trimmed and case-insensitive.
func DeviceCatalogValueKey(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}
