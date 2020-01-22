package wiino

func u64_get_byte(value uint64, shift uint8) uint8 {
    return (uint8)(value >> (shift * 8))
}

func u64_insert_byte(value uint64, shift uint8, byte uint8) uint64 {
    var mask uint64 = 0x00000000000000FF << (shift * 8)
    var inst uint64 = uint64(byte) << (shift * 8)
    return (value & ^mask) | inst
}

var table2 [8]uint8 = [8]uint8{0x1, 0x5, 0x0, 0x4, 0x2, 0x3, 0x6, 0x7}
var table1 [16]uint8 = [16]uint8{0x4, 0xB, 0x7, 0x9, 0xF, 0x1, 0xD, 0x3,
                       0xC, 0x2, 0x6, 0xE, 0x8, 0x0, 0xA, 0x5}
var table1_inv [16]uint8 = [16]uint8{0xD, 0x5, 0x9, 0x7, 0x0, 0xF, // 0xE?
                       0xA, 0x2,
                       0xC, 0x3, 0xE, 0x1, 0x8, 0x6, 0xB, 0x4}

func checkCRC(mix_id uint64) uint64 {
    var ctr int = 0
    for ctr = 0; ctr <= 42; ctr++ {
        var value uint64 = mix_id >> uint64(52 - ctr)
        if (value & 1 != 0) {
            value = 0x0000000000000635 << uint64(42 - ctr)
            mix_id ^= value
        }
        // fmt.Printf("%d ", mix_id)
    }
    // fmt.Printf("\n")
    return mix_id
}

func NWC24iMakeUserID(hollywood_id uint32, id_ctr uint16, hardware_model uint8, area_code uint8) uint64 {
    // fmt.Printf("hardware_model: %d\n", hardware_model)
    // fmt.Printf("area_code: %d\n", area_code)
    // fmt.Printf("hollywood_id: %d\n", hollywood_id)
    // fmt.Printf("id_ctr: %d\n", id_ctr)

    var mix_id uint64 = (uint64(area_code) << 50 | uint64(hardware_model) << 47 | uint64(hollywood_id) << 15 | uint64(id_ctr) << 10)
    var mix_id_copy1 uint64 = mix_id

    // fmt.Printf("7. make: %d\n", mix_id)

    mix_id = checkCRC(mix_id)

    // fmt.Printf("6. make: %d\n", mix_id)

    mix_id = (mix_id_copy1 | (mix_id & 0xFFFFFFFF)) ^ 0x0000B3B3B3B3B3B3
    mix_id = (mix_id >> 10) | ((mix_id & 0x3FF) << (11 + 32))

    // fmt.Printf("5. make: %d\n", mix_id)

    var ctr int = 0
    for ctr = 0; ctr <= 5; ctr++ {
        var ret uint8 = u64_get_byte(mix_id, uint8(ctr))
        var foobar uint8 = ((table1[(ret >> 4) & 0xF]) << 4) | (table1[ret & 0xF])
        mix_id = u64_insert_byte(mix_id, uint8(ctr), foobar & 0xff)
    }
    var mix_id_copy2 uint64 = mix_id

    // fmt.Printf("4. make: %d\n", mix_id)

    for ctr = 0; ctr <= 5; ctr++ {
        var ret uint8 = u64_get_byte(mix_id_copy2, uint8(ctr))
        mix_id = u64_insert_byte(mix_id, table2[uint8(ctr)], ret)
    }

    // fmt.Printf("3. make: %d\n", mix_id)

    mix_id &= 0x001FFFFFFFFFFFFF
    mix_id = (mix_id << 1) | ((mix_id >> 52) & 1)

    // fmt.Printf("2. make: %d\n", mix_id)

    mix_id ^= 0x00005E5E5E5E5E5E
    mix_id &= 0x001FFFFFFFFFFFFF

    // fmt.Printf("1. make: %d\n", mix_id)

    return mix_id
}

func getUnscrambleID(nwc24_id uint64) uint64 {
    var mix_id uint64 = nwc24_id

    // fmt.Printf("1. unscramble: %d\n", mix_id)

    mix_id &= 0x001FFFFFFFFFFFFF
    mix_id ^= 0x00005E5E5E5E5E5E
    mix_id &= 0x001FFFFFFFFFFFFF

    // fmt.Printf("2. unscramble: %d\n", mix_id)

    var mix_id_copy2 uint64 = mix_id

    mix_id_copy2 ^= 0xFF
    mix_id_copy2 = (mix_id << 5) & 0x20

    mix_id |= mix_id_copy2 << 48
    mix_id >>= 1

    // fmt.Printf("3. unscramble: %d\n", mix_id)

    mix_id_copy2 = mix_id

    var ctr int = 0
    for ctr = 0; ctr <= 5; ctr++ {
        var ret uint8 = u64_get_byte(mix_id_copy2, table2[ctr])
        mix_id = u64_insert_byte(mix_id, uint8(ctr), ret)
    }

    // fmt.Printf("4. unscramble: %d\n", mix_id)

    for ctr = 0; ctr <= 5; ctr++ {
        var ret uint8 = u64_get_byte(mix_id, uint8(ctr))
        var foobar uint8 = ((table1_inv[(ret >> 4) & 0xF]) << 4) | (table1_inv[ret & 0xF])
        mix_id = u64_insert_byte(mix_id, uint8(ctr), foobar & 0xff)
    }

    // fmt.Printf("5. unscramble: %d\n", mix_id)

    var mix_id_copy3 uint64 = mix_id >> 0x20
    var mix_id_copy4 uint64 = mix_id >> 0x16 | (mix_id_copy3 & 0x7FF) << 10
    var mix_id_copy5 uint64 = mix_id * 0x400 | (mix_id_copy3 >> 0xb & 0x3FF)
    var mix_id_copy6 uint64 = (mix_id_copy4 << 32) | mix_id_copy5
    var mix_id_copy7 uint64 = mix_id_copy6 ^ 0x0000B3B3B3B3B3B3
    mix_id = mix_id_copy7

    // fmt.Printf("6. unscramble: %d\n", mix_id)

    return mix_id
}

func decodeWiiID(nwc24_id uint64, hollywood_id *uint32, id_ctr *uint16, hardware_model *uint8, area_code *uint8, crc *uint16) uint64 {
    var nwc24_id2 uint64 = getUnscrambleID(nwc24_id)
    *hardware_model = uint8((nwc24_id2 >> 47) & 7)
    *area_code = uint8((nwc24_id2 >> 50) & 7)
    *hollywood_id = uint32((nwc24_id2 >> 15) & 0xFFFFFFFF)
    *id_ctr = uint16((nwc24_id2 >> 10) & 0x1F)
    *crc = uint16(nwc24_id & 0x3FF)
    // fmt.Printf("hardware_model: %d\n", *hardware_model)
    // fmt.Printf("area_code: %d\n", *area_code)
    // fmt.Printf("hollywood_id: %d\n", *hollywood_id)
    // fmt.Printf("id_ctr: %d\n", *id_ctr)
    // fmt.Printf("crc: %d\n", *crc)
    return nwc24_id2
}

func NWC24MakeUserID(hollywood_id uint32, id_ctr uint16, hardware_model uint8, area_code uint8) uint64 {
    var nwc24_id4 uint64 = NWC24iMakeUserID(hollywood_id, id_ctr, hardware_model, area_code)
    // fmt.Printf("%d", nwc24_id4)
    return nwc24_id4
}

var hollywood_id uint32
var id_ctr uint16
var hardware_model uint8
var area_code uint8
var crc uint16

func NWC24CheckUserID(nwc24_id uint64) uint8 {
    var nwc24_id3 uint64 = decodeWiiID(nwc24_id, &hollywood_id, &id_ctr, &hardware_model, &area_code, &crc)
    // fmt.Printf("%d", uint8(checkCRC(nwc24_id3)))

    return uint8(checkCRC(nwc24_id3))
}

func NWC24GetHollywoodID(nwc24_id uint64) uint32 {
    decodeWiiID(nwc24_id, &hollywood_id, &id_ctr, &hardware_model, &area_code, &crc)
    // fmt.Printf("%d", hollywood_id)

    return hollywood_id
}

func NWC24GetIDCounter(nwc24_id uint64) uint16 {
    decodeWiiID(nwc24_id, &hollywood_id, &id_ctr, &hardware_model, &area_code, &crc)
    // fmt.Printf("%d", id_ctr)

    return id_ctr
}

func NWC24GetHardwareModel(nwc24_id uint64) string {
    var models map[uint8]string = map[uint8]string{
        0: "RVT",
        1: "RVL",
        2: "RVD",
        7: "UNK",
    }

    decodeWiiID(nwc24_id, &hollywood_id, &id_ctr, &hardware_model, &area_code, &crc)
    // fmt.Printf("%d (%s)", hardware_model, models[hardware_model])

    return models[hardware_model]
}

func NWC24GetAreaCode(nwc24_id uint64) string {
    var regions map[uint8]string = map[uint8]string{
        0: "JPN",
        1: "USA",
        2: "EUR",
        3: "TWN",
        4: "KOR",
        5: "HKG",
        6: "CHN",
        7: "UNK",
    }

    decodeWiiID(nwc24_id, &hollywood_id, &id_ctr, &hardware_model, &area_code, &crc)
    // fmt.Printf("%d (%s)", hardware_model, regions[area_code])

    return regions[area_code]
}