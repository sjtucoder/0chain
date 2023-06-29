package storagesc

// Code generated by github.com/tinylib/msgp DO NOT EDIT.

import (
	"github.com/tinylib/msgp/msgp"
)

// MarshalMsg implements msgp.Marshaler
func (z *freeStorageAssigner) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 6
	// string "ClientId"
	o = append(o, 0x86, 0xa8, 0x43, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x49, 0x64)
	o = msgp.AppendString(o, z.ClientId)
	// string "PublicKey"
	o = append(o, 0xa9, 0x50, 0x75, 0x62, 0x6c, 0x69, 0x63, 0x4b, 0x65, 0x79)
	o = msgp.AppendString(o, z.PublicKey)
	// string "IndividualLimit"
	o = append(o, 0xaf, 0x49, 0x6e, 0x64, 0x69, 0x76, 0x69, 0x64, 0x75, 0x61, 0x6c, 0x4c, 0x69, 0x6d, 0x69, 0x74)
	o, err = z.IndividualLimit.MarshalMsg(o)
	if err != nil {
		err = msgp.WrapError(err, "IndividualLimit")
		return
	}
	// string "TotalLimit"
	o = append(o, 0xaa, 0x54, 0x6f, 0x74, 0x61, 0x6c, 0x4c, 0x69, 0x6d, 0x69, 0x74)
	o, err = z.TotalLimit.MarshalMsg(o)
	if err != nil {
		err = msgp.WrapError(err, "TotalLimit")
		return
	}
	// string "CurrentRedeemed"
	o = append(o, 0xaf, 0x43, 0x75, 0x72, 0x72, 0x65, 0x6e, 0x74, 0x52, 0x65, 0x64, 0x65, 0x65, 0x6d, 0x65, 0x64)
	o, err = z.CurrentRedeemed.MarshalMsg(o)
	if err != nil {
		err = msgp.WrapError(err, "CurrentRedeemed")
		return
	}
	// string "RedeemedNonces"
	o = append(o, 0xae, 0x52, 0x65, 0x64, 0x65, 0x65, 0x6d, 0x65, 0x64, 0x4e, 0x6f, 0x6e, 0x63, 0x65, 0x73)
	o = msgp.AppendArrayHeader(o, uint32(len(z.RedeemedNonces)))
	for za0001 := range z.RedeemedNonces {
		o = msgp.AppendInt64(o, z.RedeemedNonces[za0001])
	}
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *freeStorageAssigner) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var field []byte
	_ = field
	var zb0001 uint32
	zb0001, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		err = msgp.WrapError(err)
		return
	}
	for zb0001 > 0 {
		zb0001--
		field, bts, err = msgp.ReadMapKeyZC(bts)
		if err != nil {
			err = msgp.WrapError(err)
			return
		}
		switch msgp.UnsafeString(field) {
		case "ClientId":
			z.ClientId, bts, err = msgp.ReadStringBytes(bts)
			if err != nil {
				err = msgp.WrapError(err, "ClientId")
				return
			}
		case "PublicKey":
			z.PublicKey, bts, err = msgp.ReadStringBytes(bts)
			if err != nil {
				err = msgp.WrapError(err, "PublicKey")
				return
			}
		case "IndividualLimit":
			bts, err = z.IndividualLimit.UnmarshalMsg(bts)
			if err != nil {
				err = msgp.WrapError(err, "IndividualLimit")
				return
			}
		case "TotalLimit":
			bts, err = z.TotalLimit.UnmarshalMsg(bts)
			if err != nil {
				err = msgp.WrapError(err, "TotalLimit")
				return
			}
		case "CurrentRedeemed":
			bts, err = z.CurrentRedeemed.UnmarshalMsg(bts)
			if err != nil {
				err = msgp.WrapError(err, "CurrentRedeemed")
				return
			}
		case "RedeemedNonces":
			var zb0002 uint32
			zb0002, bts, err = msgp.ReadArrayHeaderBytes(bts)
			if err != nil {
				err = msgp.WrapError(err, "RedeemedNonces")
				return
			}
			if cap(z.RedeemedNonces) >= int(zb0002) {
				z.RedeemedNonces = (z.RedeemedNonces)[:zb0002]
			} else {
				z.RedeemedNonces = make([]int64, zb0002)
			}
			for za0001 := range z.RedeemedNonces {
				z.RedeemedNonces[za0001], bts, err = msgp.ReadInt64Bytes(bts)
				if err != nil {
					err = msgp.WrapError(err, "RedeemedNonces", za0001)
					return
				}
			}
		default:
			bts, err = msgp.Skip(bts)
			if err != nil {
				err = msgp.WrapError(err)
				return
			}
		}
	}
	o = bts
	return
}

// Msgsize returns an upper bound estimate of the number of bytes occupied by the serialized message
func (z *freeStorageAssigner) Msgsize() (s int) {
	s = 1 + 9 + msgp.StringPrefixSize + len(z.ClientId) + 10 + msgp.StringPrefixSize + len(z.PublicKey) + 16 + z.IndividualLimit.Msgsize() + 11 + z.TotalLimit.Msgsize() + 16 + z.CurrentRedeemed.Msgsize() + 15 + msgp.ArrayHeaderSize + (len(z.RedeemedNonces) * (msgp.Int64Size))
	return
}

// MarshalMsg implements msgp.Marshaler
func (z *freeStorageMarker) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 5
	// string "Assigner"
	o = append(o, 0x85, 0xa8, 0x41, 0x73, 0x73, 0x69, 0x67, 0x6e, 0x65, 0x72)
	o = msgp.AppendString(o, z.Assigner)
	// string "Recipient"
	o = append(o, 0xa9, 0x52, 0x65, 0x63, 0x69, 0x70, 0x69, 0x65, 0x6e, 0x74)
	o = msgp.AppendString(o, z.Recipient)
	// string "FreeTokens"
	o = append(o, 0xaa, 0x46, 0x72, 0x65, 0x65, 0x54, 0x6f, 0x6b, 0x65, 0x6e, 0x73)
	o = msgp.AppendFloat64(o, z.FreeTokens)
	// string "Nonce"
	o = append(o, 0xa5, 0x4e, 0x6f, 0x6e, 0x63, 0x65)
	o = msgp.AppendInt64(o, z.Nonce)
	// string "Signature"
	o = append(o, 0xa9, 0x53, 0x69, 0x67, 0x6e, 0x61, 0x74, 0x75, 0x72, 0x65)
	o = msgp.AppendString(o, z.Signature)
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *freeStorageMarker) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var field []byte
	_ = field
	var zb0001 uint32
	zb0001, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		err = msgp.WrapError(err)
		return
	}
	for zb0001 > 0 {
		zb0001--
		field, bts, err = msgp.ReadMapKeyZC(bts)
		if err != nil {
			err = msgp.WrapError(err)
			return
		}
		switch msgp.UnsafeString(field) {
		case "Assigner":
			z.Assigner, bts, err = msgp.ReadStringBytes(bts)
			if err != nil {
				err = msgp.WrapError(err, "Assigner")
				return
			}
		case "Recipient":
			z.Recipient, bts, err = msgp.ReadStringBytes(bts)
			if err != nil {
				err = msgp.WrapError(err, "Recipient")
				return
			}
		case "FreeTokens":
			z.FreeTokens, bts, err = msgp.ReadFloat64Bytes(bts)
			if err != nil {
				err = msgp.WrapError(err, "FreeTokens")
				return
			}
		case "Nonce":
			z.Nonce, bts, err = msgp.ReadInt64Bytes(bts)
			if err != nil {
				err = msgp.WrapError(err, "Nonce")
				return
			}
		case "Signature":
			z.Signature, bts, err = msgp.ReadStringBytes(bts)
			if err != nil {
				err = msgp.WrapError(err, "Signature")
				return
			}
		default:
			bts, err = msgp.Skip(bts)
			if err != nil {
				err = msgp.WrapError(err)
				return
			}
		}
	}
	o = bts
	return
}

// Msgsize returns an upper bound estimate of the number of bytes occupied by the serialized message
func (z *freeStorageMarker) Msgsize() (s int) {
	s = 1 + 9 + msgp.StringPrefixSize + len(z.Assigner) + 10 + msgp.StringPrefixSize + len(z.Recipient) + 11 + msgp.Float64Size + 6 + msgp.Int64Size + 10 + msgp.StringPrefixSize + len(z.Signature)
	return
}

// MarshalMsg implements msgp.Marshaler
func (z freeStorageUpgradeInput) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 2
	// string "AllocationId"
	o = append(o, 0x82, 0xac, 0x41, 0x6c, 0x6c, 0x6f, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x49, 0x64)
	o = msgp.AppendString(o, z.AllocationId)
	// string "Marker"
	o = append(o, 0xa6, 0x4d, 0x61, 0x72, 0x6b, 0x65, 0x72)
	o = msgp.AppendString(o, z.Marker)
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *freeStorageUpgradeInput) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var field []byte
	_ = field
	var zb0001 uint32
	zb0001, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		err = msgp.WrapError(err)
		return
	}
	for zb0001 > 0 {
		zb0001--
		field, bts, err = msgp.ReadMapKeyZC(bts)
		if err != nil {
			err = msgp.WrapError(err)
			return
		}
		switch msgp.UnsafeString(field) {
		case "AllocationId":
			z.AllocationId, bts, err = msgp.ReadStringBytes(bts)
			if err != nil {
				err = msgp.WrapError(err, "AllocationId")
				return
			}
		case "Marker":
			z.Marker, bts, err = msgp.ReadStringBytes(bts)
			if err != nil {
				err = msgp.WrapError(err, "Marker")
				return
			}
		default:
			bts, err = msgp.Skip(bts)
			if err != nil {
				err = msgp.WrapError(err)
				return
			}
		}
	}
	o = bts
	return
}

// Msgsize returns an upper bound estimate of the number of bytes occupied by the serialized message
func (z freeStorageUpgradeInput) Msgsize() (s int) {
	s = 1 + 13 + msgp.StringPrefixSize + len(z.AllocationId) + 7 + msgp.StringPrefixSize + len(z.Marker)
	return
}
