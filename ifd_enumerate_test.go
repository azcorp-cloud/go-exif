package exif

import (
    "path"
    "testing"
    "bytes"

    "encoding/binary"

    "github.com/dsoprea/go-logging"
)

func TestIfdTagEntry_ValueBytes(t *testing.T) {
    byteOrder := binary.BigEndian
    ve := NewValueEncoder(byteOrder)

    original := []byte("original text")

    ed, err := ve.encodeBytes(original)
    log.PanicIf(err)

    // Now, pass the raw encoded value as if it was the entire addressable area
    // and provide an offset of 0 as if it was a real block of data and this
    // value happened to be recorded at the beginning.

    ite := IfdTagEntry{
        TagType: TypeByte,
        UnitCount: uint32(len(original)),
        ValueOffset: 0,
    }

    decodedBytes, err := ite.ValueBytes(ed.Encoded, byteOrder)
    log.PanicIf(err)

    if bytes.Compare(decodedBytes, original) != 0 {
        t.Fatalf("Bytes not decoded correctly.")
    }
}

func TestIfdTagEntry_ValueBytes_RealData(t *testing.T) {
    defer func() {
        if state := recover(); state != nil {
            err := log.Wrap(state.(error))
            log.PrintErrorf(err, "Test failure.")
        }
    }()

    e := NewExif()

    filepath := path.Join(assetsPath, "NDM_8901.jpg")

    rawExif, err := e.SearchAndExtractExif(filepath)
    log.PanicIf(err)

    eh, index, err := e.Collect(rawExif)
    log.PanicIf(err)

    var ite *IfdTagEntry
    for _, thisIte := range index.RootIfd.Entries {
        if thisIte.TagId == 0x0110 {
            ite = thisIte
            break
        }
    }

    if ite == nil {
        t.Fatalf("Tag not found.")
    }

    addressableData := rawExif[ExifAddressableAreaStart:]
    decodedBytes, err := ite.ValueBytes(addressableData, eh.ByteOrder)
    log.PanicIf(err)

    expected := []byte("Canon EOS 5D Mark III")
    expected = append(expected, 0)

    if len(decodedBytes) != int(ite.UnitCount) {
        t.Fatalf("Decoded bytes not the right count.")
    } else if bytes.Compare(decodedBytes, expected) != 0 {
        t.Fatalf("Decoded bytes not correct.")
    }
}

func TestIfdTagEntry_Resolver_ValueBytes(t *testing.T) {
    defer func() {
        if state := recover(); state != nil {
            err := log.Wrap(state.(error))
            log.PrintErrorf(err, "Test failure.")
        }
    }()

    e := NewExif()

    filepath := path.Join(assetsPath, "NDM_8901.jpg")

    rawExif, err := e.SearchAndExtractExif(filepath)
    log.PanicIf(err)

    eh, index, err := e.Collect(rawExif)
    log.PanicIf(err)

    var ite *IfdTagEntry
    for _, thisIte := range index.RootIfd.Entries {
        if thisIte.TagId == 0x0110 {
            ite = thisIte
            break
        }
    }

    if ite == nil {
        t.Fatalf("Tag not found.")
    }

    itevr := NewIfdTagEntryValueResolver(rawExif, eh.ByteOrder)

    decodedBytes, err := itevr.ValueBytes(ite)
    log.PanicIf(err)

    expected := []byte("Canon EOS 5D Mark III")
    expected = append(expected, 0)

    if len(decodedBytes) != int(ite.UnitCount) {
        t.Fatalf("Decoded bytes not the right count.")
    } else if bytes.Compare(decodedBytes, expected) != 0 {
        t.Fatalf("Decoded bytes not correct.")
    }
}

func TestIfdTagEntry_Resolver_ValueBytes__Unknown_Field_And_Nonroot_Ifd(t *testing.T) {
    defer func() {
        if state := recover(); state != nil {
            err := log.Wrap(state.(error))
            log.PrintErrorf(err, "Test failure.")
        }
    }()

    e := NewExif()

    filepath := path.Join(assetsPath, "NDM_8901.jpg")

    rawExif, err := e.SearchAndExtractExif(filepath)
    log.PanicIf(err)

    eh, index, err := e.Collect(rawExif)
    log.PanicIf(err)

    ii, _ := IfdIdOrFail(IfdStandard, IfdExif)
    ifdExif := index.Lookup[ii][0]

    var ite *IfdTagEntry
    for _, thisIte := range ifdExif.Entries {
        if thisIte.TagId == 0x9000 {
            ite = thisIte
            break
        }
    }

    if ite == nil {
        t.Fatalf("Tag not found.")
    }

    itevr := NewIfdTagEntryValueResolver(rawExif, eh.ByteOrder)

    decodedBytes, err := itevr.ValueBytes(ite)
    log.PanicIf(err)

    expected := []byte { '0', '2', '3', '0' }

    if len(decodedBytes) != int(ite.UnitCount) {
        t.Fatalf("Decoded bytes not the right count.")
    } else if bytes.Compare(decodedBytes, expected) != 0 {
        t.Fatalf("Recovered unknown value is not correct.")
    }
}

func Test_Ifd_FindTagWithId_Hit(t *testing.T) {
    filepath := path.Join(assetsPath, "NDM_8901.jpg")

    e := NewExif()

    rawExif, err := e.SearchAndExtractExif(filepath)
    log.PanicIf(err)

    _, index, err := e.Collect(rawExif)
    log.PanicIf(err)

    ifd := index.RootIfd
    results, err := ifd.FindTagWithId(0x011b)

    if len(results) != 1 {
        t.Fatalf("Exactly one result was not found: (%d)", len(results))
    } else if results[0].TagId != 0x011b {
        t.Fatalf("The result was not expected: %v", results[0])
    }
}

func Test_Ifd_FindTagWithId_Miss(t *testing.T) {
    filepath := path.Join(assetsPath, "NDM_8901.jpg")

    e := NewExif()

    rawExif, err := e.SearchAndExtractExif(filepath)
    log.PanicIf(err)

    _, index, err := e.Collect(rawExif)
    log.PanicIf(err)

    ifd := index.RootIfd

    _, err = ifd.FindTagWithId(0xffff)
    if err == nil {
        t.Fatalf("Expected error for not-found tag.")
    } else if log.Is(err, ErrTagNotFound) == false {
        log.Panic(err)
    }
}

func Test_Ifd_FindTagWithName_Hit(t *testing.T) {
    filepath := path.Join(assetsPath, "NDM_8901.jpg")

    e := NewExif()

    rawExif, err := e.SearchAndExtractExif(filepath)
    log.PanicIf(err)

    _, index, err := e.Collect(rawExif)
    log.PanicIf(err)

    ifd := index.RootIfd
    results, err := ifd.FindTagWithName("YResolution")

    if len(results) != 1 {
        t.Fatalf("Exactly one result was not found: (%d)", len(results))
    } else if results[0].TagId != 0x011b {
        t.Fatalf("The result was not expected: %v", results[0])
    }
}

func Test_Ifd_FindTagWithName_Miss(t *testing.T) {
    filepath := path.Join(assetsPath, "NDM_8901.jpg")

    e := NewExif()

    rawExif, err := e.SearchAndExtractExif(filepath)
    log.PanicIf(err)

    _, index, err := e.Collect(rawExif)
    log.PanicIf(err)

    ifd := index.RootIfd

    _, err = ifd.FindTagWithName("PlanarConfiguration")
    if err == nil {
        t.Fatalf("Expected error for not-found tag.")
    } else if log.Is(err, ErrTagNotFound) == false {
        log.Panic(err)
    }
}

func Test_Ifd_FindTagWithName_NonStandard(t *testing.T) {
    filepath := path.Join(assetsPath, "NDM_8901.jpg")

    e := NewExif()

    rawExif, err := e.SearchAndExtractExif(filepath)
    log.PanicIf(err)

    _, index, err := e.Collect(rawExif)
    log.PanicIf(err)

    ifd := index.RootIfd

    _, err = ifd.FindTagWithName("GeorgeNotAtHome")
    if err == nil {
        t.Fatalf("Expected error for not-found tag.")
    } else if log.Is(err, ErrTagNotStandard) == false {
        log.Panic(err)
    }
}