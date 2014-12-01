// Parses files containing server descriptors.

package zoossh

import (
	"io/ioutil"
	"strconv"
	"strings"
	"time"
)

const (
	// The last line of a router descriptor.
	descriptorDelimiter string = "-----END SIGNATURE-----"
	// The layout of the "published" field.
	publishedTimeLayout string = "2006-01-02 15:04:05"
)

// An exitpattern as defined in dirspec.txt, Section 2.1.3.
type ExitPattern struct {
	AddressSpec string
	PortSpec    string
}

// An (incomplete) router descriptor as defined in dirspec.txt, Section 2.1.1.
type RouterDescriptor struct {

	// The single fields of a "router" line.
	Nickname  string
	Address   string
	ORPort    uint16
	SOCKSPort uint16
	DirPort   uint16

	// The single fields of a "bandwidth" line.  All bandwidth values are in
	// bytes per second.
	BandwidthAvg   uint64
	BandwidthBurst uint64
	BandwidthObs   uint64

	// The single fields of a "platform" line.
	OperatingSystem string
	TorVersion      string

	// The single fields of a "published" line.
	Published time.Time

	// The single fields of an "uptime" line.
	Uptime uint64

	// The single fields of a "fingerprint" line.
	Fingerprint string

	// The single fields of a "hibernating" line.
	Hibernating bool

	// The single fields of a "family" line.
	Family []string

	// The single fields of a "contact" line.
	Contact string

	// The "hidden-service-dir" line.
	HiddenServiceDir bool

	OnionKey     string
	NTorOnionKey string
	SigningKey   string

	Accept []*ExitPattern
	Reject []*ExitPattern
}

func ParseDescriptorString(rawDescriptor string) (*RouterDescriptor, error) {

	var descriptor *RouterDescriptor = new(RouterDescriptor)
	var port uint64

	lines := strings.Split(rawDescriptor, "\n")

	// Go over raw descriptor line by line and extract the fields we are
	// interested in.
	for _, line := range lines {

		words := strings.Split(line, " ")

		switch words[0] {

		case "router":
			descriptor.Nickname = words[1]
			descriptor.Address = words[2]
			port, _ = strconv.ParseUint(words[3], 10, 16)
			descriptor.ORPort = uint16(port)
			port, _ = strconv.ParseUint(words[4], 10, 16)
			descriptor.SOCKSPort = uint16(port)
			port, _ = strconv.ParseUint(words[5], 10, 16)
			descriptor.DirPort = uint16(port)

		case "platform":
			descriptor.OperatingSystem = words[4]
			descriptor.TorVersion = words[2]

		case "uptime":
			descriptor.Uptime, _ = strconv.ParseUint(words[1], 10, 64)

		case "published":
			time, _ := time.Parse(publishedTimeLayout, strings.Join(words[1:], " "))
			descriptor.Published = time

		case "fingerprint":
			descriptor.Fingerprint = strings.Join(words[1:], "")

		case "hibernating":
			descriptor.Hibernating, _ = strconv.ParseBool(words[1])

		case "bandwidth":
			descriptor.BandwidthAvg, _ = strconv.ParseUint(words[1], 10, 64)
			descriptor.BandwidthBurst, _ = strconv.ParseUint(words[2], 10, 64)
			descriptor.BandwidthObs, _ = strconv.ParseUint(words[3], 10, 64)

		case "family":
			descriptor.Family = words[1:]

		case "contact":
			descriptor.Contact = strings.Join(words[1:], " ")

		case "hidden-service-dir":
			descriptor.HiddenServiceDir = true
		}
	}

	return descriptor, nil
}

func ParseDescriptorFile(fileName string) ([]RouterDescriptor, error) {

	blurb, err := ioutil.ReadFile(fileName)
	if err != nil {
		return nil, err
	}

	var rawDescs string = string(blurb)
	var descriptors []RouterDescriptor

	for {
		// Jump to the end of the next relay descriptor.
		position := strings.Index(rawDescs, descriptorDelimiter)
		if position == -1 {
			break
		}
		position += len(descriptorDelimiter)

		// Turn raw textual descriptor into struct.
		descriptor, err := ParseDescriptorString(rawDescs[:position])
		if err != nil {
			return nil, err
		}
		descriptors = append(descriptors, *descriptor)

		// Point to the beginning of the next relay descriptor.
		rawDescs = rawDescs[position:]
	}

	return descriptors, nil
}
