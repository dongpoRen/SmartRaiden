package attr

import (
	"encoding/binary"
	"strconv"

	"github.com/SmartMeshFoundation/SmartRaiden/network/nat/goice/stun"
	"github.com/SmartMeshFoundation/SmartRaiden/network/nat/goice/turn"
)

type IceControlled uint64

func (n IceControlled) String() string { return strconv.Itoa(int(n)) }

const IceControlledSize = 8

// AddTo adds ICE-CONTROLLED  to message.
func (n IceControlled) AddTo(m *stun.Message) error {
	v := make([]byte, IceControlledSize)
	binary.BigEndian.PutUint64(v, uint64(n))
	m.Add(stun.AttrICEControlled, v)
	return nil
}

// GetFrom decodes ICE-CONTROLLED from message.
func (n *IceControlled) GetFrom(m *stun.Message) error {
	v, err := m.Get(stun.AttrICEControlled)
	if err != nil {
		return err
	}
	if len(v) != IceControlledSize {
		return &turn.BadAttrLength{
			Attr:     stun.AttrICEControlled,
			Got:      len(v),
			Expected: IceControlledSize,
		}
	}
	*n = IceControlled(binary.BigEndian.Uint64(v))
	return nil
}
