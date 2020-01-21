package main

import "fmt"

func u64_get_byte(value uint64, shift uint8) uint8 {
    return (uint8)(value >> (shift * 8))
}

func u64_insert_byte(value uint64, shift uint8, byte uint8) uint64 {
    var mask uint64 = 0x00000000000000FF << (shift * 8)
    var inst uint64 = uint64(byte) << (shift * 8)
    return (value & ^mask) | inst
}

var table2 [8]uint8 = [8]uint8{0x1, 0x5, 0x0, 0x4, 0x2, 0x3, 0x6, 0x7};
var table1 [16]uint8 = [16]uint8{0x4, 0xB, 0x7, 0x9, 0xF, 0x1, 0xD, 0x3,
                       0xC, 0x2, 0x6, 0xE, 0x8, 0x0, 0xA, 0x5};
var table1_inv [16]uint8 = [16]uint8{0xD, 0x5, 0x9, 0x7, 0x0, 0xF, // 0xE?
                       0xA, 0x2,
                       0xC, 0x3, 0xE, 0x1, 0x8, 0x6, 0xB, 0x4};

func getUnscrambleID(nwc24_id uint64) uint64 {
    var mix_id uint64 = nwc24_id

    fmt.Printf("1. unscramble: %d\n", mix_id)

    mix_id &= 0x001FFFFFFFFFFFFF
    mix_id ^= 0x00005E5E5E5E5E5E
    mix_id &= 0x001FFFFFFFFFFFFF

    fmt.Printf("2. unscramble: %d\n", mix_id)

    var mix_id_copy2 uint64 = mix_id

    mix_id_copy2 ^= 0xFF
    mix_id_copy2 = (mix_id << 5) & 0x20

    mix_id |= mix_id_copy2 << 48
    mix_id >>= 1

    fmt.Printf("3. unscramble: %d\n", mix_id)

    mix_id_copy2 = mix_id

    var ctr int = 0
    for ctr = 0; ctr <= 5; ctr++ {
        var ret uint8 = u64_get_byte(mix_id_copy2, table2[ctr])
        mix_id = u64_insert_byte(mix_id, uint8(ctr), ret)
    }

    fmt.Printf("4. unscramble: %d\n", mix_id)

    for ctr = 0; ctr <= 5; ctr++ {
        var ret uint8 = u64_get_byte(mix_id, uint8(ctr))
        var foobar uint8 = ((table1_inv[(ret >> 4) & 0xF]) << 4) | (table1_inv[ret & 0xF])
        mix_id = u64_insert_byte(mix_id, uint8(ctr), foobar & 0xff)
    }

    fmt.Printf("5. unscramble: %d\n", mix_id)

    var mix_id_copy3 uint64 = mix_id >> 0x20
    var mix_id_copy4 uint64 = mix_id >> 0x16 | (mix_id_copy3 & 0x7FF) << 10
    var mix_id_copy5 uint64 = mix_id * 0x400 | (mix_id_copy3 >> 0xb & 0x3FF)
    var mix_id_copy6 uint64 = (mix_id_copy4 << 32) | mix_id_copy5
    var mix_id_copy7 uint64 = mix_id_copy6 ^ 0x0000B3B3B3B3B3B3
    mix_id = mix_id_copy7

    fmt.Printf("6. unscramble: %d\n", mix_id)

    return mix_id
}

func main() {
    getUnscrambleID(6330930957365086)
}