#include "wiino.h"

#include <map>
#include <string>

u8 GetAreaCode(const std::string& area)
{
  static const std::map<std::string, u8> regions = {
      {"JPN", 0}, {"USA", 1}, {"EUR", 2}, {"AUS", 2}, {"BRA", 1}, {"TWN", 3}, {"ROC", 3},
      {"KOR", 4}, {"HKG", 5}, {"ASI", 5}, {"LTN", 1}, {"SAF", 2}, {"CHN", 6},
  };

  auto entryPos = regions.find(area);
  if (entryPos != regions.end())
    return entryPos->second;

  return 7;  // Unknown
}

u8 GetHardwareModel(const std::string& model)
{
  static const std::map<std::string, u8> models = {
      {"RVL", 1},
      {"RVT", 0},
      {"RVV", 0},
      {"RVD", 2},
  };

  auto entryPos = models.find(model);
  if (entryPos != models.end())
    return entryPos->second;

  return 7;
}

static u8 u64_get_byte(u64 value, u8 shift)
{
  return (u8)(value >> (shift * 8));
}

static u64 u64_insert_byte(u64 value, u8 shift, u8 byte)
{
  u64 mask = 0x00000000000000FFULL << (shift * 8);
  u64 inst = (u64)byte << (shift * 8);
  return (value & ~mask) | inst;
}

static u8 u64_get_byte2(u64 value, u8 byte)
{
    return (value >> (byte << 3)) & 0xFF;
}

static u64 u64_insert_byte2(u64 value, u8 newval, u8 byte)
{
    return (value & (0xFF << (byte << 3))) | (newval << (byte << 3));
}

const u8 table2[8] = {0x1, 0x5, 0x0, 0x4, 0x2, 0x3, 0x6, 0x7};
const u8 table1[16] = {0x4, 0xB, 0x7, 0x9, 0xF, 0x1, 0xD, 0x3,
                       0xC, 0x2, 0x6, 0xE, 0x8, 0x0, 0xA, 0x5};
const u8 table1_inv[16] = {0xD, 0x5, 0x9, 0x7, 0x0, 0xF, // 0xE?
                       0xA, 0x2,
                       0xC, 0x3, 0xE, 0x1, 0x8, 0x6, 0xB, 0x4};

s32 NWC24MakeUserID(u64* nwc24_id, u32 hollywood_id, u16 id_ctr, u8 hardware_model,
                                  u8 area_code)
{
    printf("hardware_model: %u\n", hardware_model);
    printf("area_code: %u\n", area_code);
    printf("hollywood_id: %u\n", hollywood_id);
    printf("id_ctr: %u\n", id_ctr);

    u64 mix_id = ((u64)area_code << 50) | ((u64)hardware_model << 47) | ((u64)hollywood_id << 15) |
                    ((u64)id_ctr << 10);

    printf("7. make: %llu\n", mix_id);

    int ctr = 0;
    for (ctr = 0; ctr <= 42; ctr++)
    {
        u64 value = mix_id >> (52 - ctr);
        if (value & 1)
        {
            value = 0x0000000000000635ULL << (42 - ctr);
            mix_id ^= value;
        }
        printf("%u ", mix_id);
    }
    printf("\n");

    printf("6. make: %llu\n", mix_id);

    u64 mix_id_copy1 = mix_id;

    mix_id = (mix_id_copy1 | (mix_id & 0xFFFFFFFFUL)) ^ 0x0000B3B3B3B3B3B3ULL;
    mix_id = (mix_id >> 10) | ((mix_id & 0x3FF) << (11 + 32));

    printf("5. make: %llu\n", mix_id);

    for (ctr = 0; ctr <= 5; ctr++)
    {
    u8 ret = u64_get_byte(mix_id, ctr);
    u8 foobar = ((table1[(ret >> 4) & 0xF]) << 4) | (table1[ret & 0xF]);
    mix_id = u64_insert_byte(mix_id, ctr, foobar & 0xff);
    }
    u64 mix_id_copy2 = mix_id;

    printf("4. make: %llu\n", mix_id);

    for (ctr = 0; ctr <= 5; ctr++)
    {
    u8 ret = u64_get_byte(mix_id_copy2, ctr);
    mix_id = u64_insert_byte(mix_id, table2[ctr], ret);
    }

    printf("3. make: %llu\n", mix_id);

    mix_id &= 0x001FFFFFFFFFFFFFULL;
    mix_id = (mix_id << 1) | ((mix_id >> 52) & 1);

    printf("2. make: %llu\n", mix_id);

    mix_id ^= 0x00005E5E5E5E5E5EULL;
    mix_id &= 0x001FFFFFFFFFFFFFULL;

    printf("1. make: %llu\n", mix_id);

    *nwc24_id = mix_id;

    if (mix_id > 9999999999999999ULL)
    return 1;

    return 0;
}

u64 getUnScrambleId(u64 nwc24_id)
{
    u64 mix_id = nwc24_id;
    u8 a;
    u8 b;

    printf("1. unscramble: %llu\n", mix_id);

    mix_id &= 0x001FFFFFFFFFFFFFULL;
    mix_id ^= 0x00005E5E5E5E5E5EULL;
    mix_id &= 0x001FFFFFFFFFFFFFULL;

    printf("2. unscramble: %llu\n", mix_id);

    u64 mix_id_copy2 = mix_id;

    mix_id_copy2 ^= 0xFF;
    mix_id_copy2 = (mix_id << 5) & 0x20;

    mix_id |= mix_id_copy2 << 48;
    mix_id >>= 1;

    printf("3. unscramble: %llu\n", mix_id);

    mix_id_copy2 = mix_id;

    int ctr = 0;
    for (ctr = 0; ctr <= 5; ctr++)
    {
        u8 ret = u64_get_byte(mix_id_copy2, table2[ctr]);
        mix_id = u64_insert_byte(mix_id, ctr, ret);
    }

    printf("4. unscramble: %llu\n", mix_id);

    for (ctr = 0; ctr <= 5; ctr++)
    {
        u8 ret = u64_get_byte(mix_id, ctr);
        u8 foobar = ((table1_inv[(ret >> 4) & 0xF]) << 4) | (table1_inv[ret & 0xF]);
        mix_id = u64_insert_byte(mix_id, ctr, foobar & 0xff);
    }

    printf("5. unscramble: %llu\n", mix_id);

    u64 mix_id_copy3 = mix_id >> 0x20;
    u64 mix_id_copy4 = mix_id >> 0x16 | (mix_id_copy3 & 0x7FF) << 10;
    u64 mix_id_copy5 = mix_id * 0x400 | mix_id_copy3 >> 0xb & 0x3FF;
    u64 mix_id_copy6 = (mix_id_copy4 << 32) | mix_id_copy5;
    u64 mix_id_copy7 = mix_id_copy6 ^ 0x0000B3B3B3B3B3B3ULL;
    u16 unused = mix_id_copy7 & 0x3FF;

    printf("6. unscramble: %llu\n", mix_id_copy7);

    return mix_id;
}

u64 decodeWiiId(u64 nwc24_id, u32 *hollywood_id, u16 *id_ctr, u8 *hardware_model, u8 *area_code, u16 *unused)
{
    u64 nwc24_id2 = getUnScrambleId(nwc24_id);
    *hardware_model = (nwc24_id2 >> 47) & 7;
    *area_code = (nwc24_id2 >> 50) & 7;
    *hollywood_id = (nwc24_id2 >> 15) & 0xFFFFFFFF;
    *id_ctr = (nwc24_id2 >> 10) & 0x1F;
    printf("hardware_model: %u\n", hardware_model);
    printf("area_code: %u\n", area_code);
    printf("hollywood_id: %u\n", hollywood_id);
    printf("id_ctr: %u\n", id_ctr);
    return nwc24_id;
}

s32 CheckUserIdInternal(u64 check_id)
{
    u8 cRegion, cModel;
    u16 cUnused, cCounter;
    u32 cHollywoodId;
    u64 crc2 = decodeWiiId(check_id, &cHollywoodId, &cCounter, &cModel, &cRegion, &cUnused);
    return 0;
}

s32 NWC24CheckUserId(u64 nwc24_id)
{
    CheckUserIdInternal(nwc24_id);
}

s32 main()
{
    u64 nwc24_id;
    NWC24MakeUserID(&nwc24_id, 2, 1, GetHardwareModel("RVL"), GetAreaCode("USA"));
    printf("user id generated: %llu\n", nwc24_id);
    NWC24CheckUserId(nwc24_id);
}