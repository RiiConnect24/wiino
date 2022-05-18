def u64_get_byte(value, shift):
    byte = (value >> (shift * 8)) & 0xff
    return byte

def u64_insert_byte(value, shift, byte):
    mask = 0x00000000000000FF << (shift * 8)
    inst = byte << (shift * 8)
    return (value & ~mask) | inst
    
table2 = [0x1, 0x5, 0x0, 0x4, 0x2, 0x3, 0x6, 0x7]
table1 = [0x4, 0xB, 0x7, 0x9, 0xF, 0x1, 0xD, 0x3, 0xC, 0x2, 0x6, 0xE, 0x8, 0x0, 0xA, 0x5]
table1_inv = [0xD, 0x5, 0x9, 0x7, 0x0, 0xF, 0xA, 0x2, 0xC, 0x3, 0xE, 0x1, 0x8, 0x6, 0xB, 0x4]

def checkCRC(mix_id):
    for ctr in range(0, 43):
        value = mix_id >> (52 - ctr)
        if value & 1:
            value = 0x0000000000000635 << (42 - ctr)
            mix_id ^= value
    return mix_id

def NWC24iMakeUserID(hollywood_id, id_ctr, hardware_model, area_code):
    mix_id = (area_code << 50) | (hardware_model << 47) | (hollywood_id << 15) | (id_ctr << 10);
    mix_id_copy1 = mix_id
    
    mix_id = checkCRC(mix_id)
    
    mix_id = (mix_id_copy1 | (mix_id & 0xFFFFFFFF)) ^ 0x0000B3B3B3B3B3B3
    mix_id = (mix_id >> 10) | ((mix_id & 0x3FF) << (11 + 32))
    
    for ctr in range(0, 6):
        ret = u64_get_byte(mix_id, ctr)
        foobar = ((table1[(ret >> 4) & 0xF]) << 4) | (table1[ret & 0xF])
        mix_id = u64_insert_byte(mix_id, ctr, foobar & 0xff)
        
    mix_id_copy2 = mix_id
    
    for ctr in range(0, 6):
        ret = u64_get_byte(mix_id_copy2, ctr)
        mix_id = u64_insert_byte(mix_id, table2[ctr], ret)
        
    mix_id &= 0x001FFFFFFFFFFFFF
    mix_id = (mix_id << 1) | ((mix_id >> 52) & 1)

    mix_id ^= 0x00005E5E5E5E5E5E
    mix_id &= 0x001FFFFFFFFFFFFF
    return mix_id

def getUnScrambleID(mix_id):
    mix_id &= 0x001FFFFFFFFFFFFF
    mix_id ^= 0x00005E5E5E5E5E5E
    mix_id &= 0x001FFFFFFFFFFFFF
    
    mix_id_copy2 = (mix_id << 5) & 0x20

    mix_id |= mix_id_copy2 << 48
    mix_id >>= 1

    mix_id_copy2 = mix_id

    for ctr in range(0, 6):
        ret = u64_get_byte(mix_id_copy2, table2[ctr])
        mix_id = u64_insert_byte(mix_id, ctr, ret)

    for ctr in range(0, 6):
        ret = u64_get_byte(mix_id, ctr)
        foobar = ((table1_inv[(ret >> 4) & 0xF]) << 4) | (table1_inv[ret & 0xF])
        mix_id = u64_insert_byte(mix_id, ctr, foobar & 0xff)

    mix_id_copy3 = mix_id >> 0x20
    mix_id_copy4 = mix_id >> 0x16 | (mix_id_copy3 & 0x7FF) << 10
    mix_id_copy5 = mix_id * 0x400 | (mix_id_copy3 >> 0xb & 0x3FF)
    mix_id_copy6 = (mix_id_copy4 << 32) | mix_id_copy5
    mix_id_copy7 = mix_id_copy6 ^ 0x0000B3B3B3B3B3B3
    mix_id = mix_id_copy7

    return mix_id

def decodeWiiID(nwc24_id):
    nwc24_id2 = getUnScrambleID(nwc24_id)
    hardware_model = (nwc24_id2 >> 47) & 7
    area_code = (nwc24_id2 >> 50) & 7
    hollywood_id = (nwc24_id2 >> 15) & 0xFFFFFFFF
    id_ctr = (nwc24_id2 >> 10) & 0x1F
    crc = nwc24_id2 & 0x3FF
    return (nwc24_id2, hardware_model, area_code, hollywood_id, id_ctr, crc)
    
def NWC24CheckUserID(nwc24_id):
    unscrambled = decodeWiiID(nwc24_id)[0]
    return checkCRC(unscrambled) & 0xFFFFFFFF

def NWC24GetHollywoodID(nwc24_id):
    hollywood = decodeWiiID(nwc24_id)[3]
    return hollywood

def NWC24GetIDCounter(nwc24_id):
    id_ctr = decodeWiiID(nwc24_id)[4]
    return id_ctr
    
def NWC24GetHardwareModel(nwc24_id):
    hardware_model = decodeWiiID(nwc24_id)[1]
    return hardware_model
    
def NWC24GetAreaCode(nwc24_id):
    area_code = decodeWiiID(nwc24_id)[2]
    return area_code
