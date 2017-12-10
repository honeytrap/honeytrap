package farm

import "testing"

var testData = []struct {
	oh32, oh32s                 uint32
	oh64na, oh64, oh64s, oh64ss uint64
	oh128lo, oh128hi            uint64
	oh128slo, oh128shi          uint64
	in                          string
}{
	{0xdc56d17a, 0x0108292b, 0x9ae16a3b2f90404f, 0x9ae16a3b2f90404f, 0xb0403333574d37e4, 0x1ad9361c3f563461, 0x3df09dfc64c09a2b, 0x3cb540c392e51e29, 0x9fd4b80883017649, 0x806308c81d07d094, ""},
	{0x3c973d4d, 0x7e4cfeed, 0xb3454265b6df75e3, 0xb3454265b6df75e3, 0x779ef0ca4870bcc2, 0xd95bffbe1f68b36d, 0x6e97d6bbdfc0a0c4, 0x52a71e38f43be561, 0xa347ea476dd92aff, 0xe12da4d2563e7840, "a"},
	{0x417330fd, 0x546fd574, 0xaa8d6e5242ada51e, 0xaa8d6e5242ada51e, 0xbfb091da4ffef5d8, 0xa05bfc8f958ce312, 0x13e834f38a6c88b8, 0xcfdbce01c0e7622e, 0x4cfab1d0801d789d, 0x1cb9185cd5d6c743, "ab"},
	{0x2f635ec7, 0x27b6c746, 0x24a5b3a074e7f369, 0x24a5b3a074e7f369, 0x7c6dc4691a7576b6, 0x008b6aa9b6344c8d, 0x3980b2afd2126c04, 0xa085f09013029e45, 0xbeadc73cd0b92afe, 0xd0d697a5a943a657, "abc"},
	{0x98b51e95, 0xdfb26aae, 0x1a5502de4a1f8101, 0x1a5502de4a1f8101, 0xfdc6b0ac6369ca3a, 0xf0d7f476a4cc1a4a, 0xb8d7175e11647e82, 0x0906d778016538d9, 0x83799d5bc388066c, 0x81f582de4830d3ef, "abcd"},
	{0xa3f366ac, 0xd29c0f4d, 0xc22f4663e54e04d4, 0xc22f4663e54e04d4, 0x64b5b351c1cee4a2, 0x6ecf50380f033790, 0x940fcbbc468d384f, 0xf7776b2eaa1583e1, 0x752e188459fdb79b, 0xaccbe81c5b1e610a, "abcde"},
	{0x0f813aa4, 0x4abfab6e, 0xc329379e6a03c2cd, 0xc329379e6a03c2cd, 0xb8e8ed73451fc99f, 0x03850d708bfcb24b, 0x6f7c444b0a4eb3eb, 0x7da95bbe683b00b0, 0x09ca9b9287f397d9, 0x0dc7ad8161b168d6, "abcdef"},
	{0x21deb6d7, 0xbd18c595, 0x3c40c92b1ccb7355, 0x3c40c92b1ccb7355, 0x66301d1183086f9c, 0x25fcd92789893a23, 0x9e5daa7baf7e4573, 0xcfb5d54dd0ac6959, 0x45ada3ef413e2174, 0xe05f1b93e468b6df, "abcdefg"},
	{0xfd7ec8b9, 0x759ee8b6, 0xfee9d22990c82909, 0xfee9d22990c82909, 0xef6b24f6a776fd30, 0xee098adfe88e7340, 0x60f2a826d4d614ef, 0x56f19716a4032fcb, 0x6e315acf85d006a2, 0xdc645ba0560c2dc9, "abcdefgh"},
	{0x6f98dc86, 0xa7440120, 0x332c8ed4dae5ba42, 0x332c8ed4dae5ba42, 0xb54b4c8bd2372ffc, 0x4c405cec37506bfd, 0x7b5fd93102612d91, 0x61aa40e4e386bd9c, 0xf6120c3bbe5f4aff, 0x7d2f90ca09078fa8, "abcdefghi"},
	{0x9741ca1a, 0xdb4c64d2, 0xad052244b781c4eb, 0xad052244b781c4eb, 0xd56f3e3e18d9cacc, 0xe148f198dad5079d, 0xcf650cdfad0d0675, 0x6f6cb496e4b429c5, 0x5f9889eedca799d1, 0xee9890e8b8567c89, "0123456789"},
	{0xca179ba9, 0x7fca6fd6, 0x3ef4c03514208c77, 0x3ef4c03514208c77, 0xda0851051d8a9854, 0xdd0bdad27bda1972, 0xc82790fefc8d709e, 0x95fac84d6ce3f311, 0x000153a7819942df, 0x8ee8ba5010b1da1c, "0123456789 "},
	{0xf8cc7928, 0x00be5c31, 0x496841e83a33cc91, 0x496841e83a33cc91, 0x27f085304094b99f, 0x0e85e1954a10b28e, 0x334199771269f58b, 0xdfe28ae349106abd, 0x2231c7f1bfd9b5d6, 0x50a4d4015221e228, "0123456789-0"},
	{0x0d92cafb, 0xe6588500, 0xd81bcb9f3679ac0c, 0xd81bcb9f3679ac0c, 0x938a277ef2446d80, 0x378ce8623423a232, 0x534f80dc46200e5b, 0x33de7f2feb6e9f9c, 0x9b8e156e09691d6a, 0xe80b02c1357f2c45, "0123456789~01"},
	{0x71a36842, 0xe514ed82, 0x5da5a6a117c606f6, 0x5da5a6a117c606f6, 0x2439b47b63cfb868, 0x9b0a3f5504372975, 0x4ffccef1542c2681, 0x481bec885727f698, 0x7b31a7b9c6950bce, 0x4feef06f8e8343f5, "0123456789#012"},
	{0x93ee6801, 0x0cfc86d2, 0x5361eae17c1ff6bc, 0x5361eae17c1ff6bc, 0xf629b20de846fc23, 0x3ca6d68cb0129d5f, 0x8ea7f4ceea677800, 0x27f04b052a82a69d, 0xc2bc043bcda6164e, 0xf11906ca790e0d90, "0123456789@0123"},
	{0x9cecd750, 0x93278a98, 0x4283d4ef43627f64, 0x4283d4ef43627f64, 0x33abb5361a4815c1, 0x5bec11b7f0eb5b0e, 0xbb5db1cf64974657, 0x620437956a2c3feb, 0x1b5d8f018909205d, 0xc1b3ecfee087c7af, "0123456789'01234"},
	{0x335f081f, 0x34981478, 0x46a7416ed4861e3b, 0x46a7416ed4861e3b, 0xf4547ea9e910bc0e, 0x5cfaf00ae1d9b34c, 0x18814368ecc30fa6, 0xa73eb969f0303770, 0x3dbd3fa7a545bf08, 0x991d50420e35a9ff, "0123456789=012345"},
	{0xa9785062, 0x7580b9fb, 0xa4abb4e0da2c594c, 0xa4abb4e0da2c594c, 0x05d865e3844b83fe, 0x8a6aa12f6f761cf2, 0xb86e8b3e907ea7f9, 0xd28023bc18e339ef, 0x77c6ecb4928fd5e7, 0xe8b55ddbfef07956, "0123456789+0123456"},
	{0x5d4bd7f6, 0x73bdc9fb, 0xcf1c7d3ad54f9215, 0xcf1c7d3ad54f9215, 0x79be62145d9d67cf, 0x6541cb3eebafd693, 0xc83a8801c1057f7a, 0x7a6ef6954512e51b, 0x49daf680f01b4e79, 0x42eeb6d63a4a2481, "0123456789*01234567"},
	{0x3884aa05, 0xda41f50b, 0x07adf50b2ac764fc, 0x07adf50b2ac764fc, 0x2474703a1b29f6a9, 0xb01e016d86ede665, 0x7f14c7b52d596e44, 0x6767246d19bc4e34, 0x3d154e894a20b400, 0x76e45ed353112875, "0123456789&012345678"},
	{0x536d1efd, 0xcbd6cb47, 0xdebcba8e6f3eabd1, 0xdebcba8e6f3eabd1, 0x9fb28e096e6968db, 0x91b2d71a5147967b, 0x321c9c8e18ea81f4, 0xc4acf23d0fbc5484, 0xc4c32b477c0574dc, 0x9440e0adeba7673b, "0123456789^0123456789"},
	{0x1723dd7a, 0xf3485759, 0x4dbd128af51d77e8, 0x4dbd128af51d77e8, 0x1e9ce746b2ffef52, 0xac60f1c73ac2c016, 0x6e80a3ba73041556, 0x0a5d94eaead48b58, 0x64a7728894381e6b, 0x71d9d110555d07b2, "0123456789%0123456789£"},
	{0xfa88d020, 0xd89da44b, 0xd78d5f852d522e6a, 0xd78d5f852d522e6a, 0x072fd1a4b21ed852, 0x7a7ddf0143013fc2, 0xe91658c980e0f179, 0x3eb8d05028d2fe3c, 0xcaf33ff7add6b07b, 0xfcc2de2d2f0f458b, "0123456789$0123456789!0"},
	{0xc6246b8d, 0x9f99edf0, 0x80d73b843ba57db8, 0x80d73b843ba57db8, 0x8d8d5757a7bda0ba, 0xb9857c5b86ba9e86, 0x12ab5d9fca8a7a6f, 0x26726f42f1aba3b3, 0x33bb4a26e209e9f8, 0x00f453ad18334e87, "size:  a.out:  bad magic"},
	{0x322984d9, 0xb630933e, 0x8eb3808d1ccfc779, 0x8eb3808d1ccfc779, 0x5ba1410bdca911a7, 0x78fc99025fd288e6, 0xa646f296be6c7a80, 0xc5c0c296fea38db2, 0x481512913a7eb34a, 0xa6d72dee130312f8, "Nepal premier won't resign."},
	{0x221694e4, 0xb959719f, 0xb944f8a16261e414, 0xb944f8a16261e414, 0x4e4d65c93956a760, 0x5edcdee0fa2d5936, 0x1e41fc4638c4da77, 0x0896f4bd73582b3e, 0xb7a0568d25721aa1, 0xdbb71e5308b97e38, "C is as portable as Stonehedge!!"},
	{0xe273108f, 0x6d328965, 0xe8f89ab6df9bdd25, 0xe8f89ab6df9bdd25, 0xa3a2a2a6c80ebbd2, 0xf88dbcbab3c05f4e, 0xc6c8eac0aafacfed, 0x8efcd3bd44573235, 0x445f11917c9e20cd, 0x6b22d6c3239d923d, "Discard medicine more than two years old."},
	{0x363394d1, 0xc5010f28, 0xa9961670ce2a46d9, 0xa9961670ce2a46d9, 0x9d0526b9480738b8, 0xf31be505a4ee2623, 0xea8b15a2a33e8211, 0xe23fc1fda1552993, 0xb589f51ad7354d27, 0x2322b93f6ce62132, "I wouldn't marry him with a ten foot pole."},
	{0x4fda5f07, 0x40172aa7, 0xbdd69b798d6ba37a, 0xbdd69b798d6ba37a, 0x1110fe437d57c4d3, 0x74a9361763b11819, 0x5595b0dbcb471e00, 0xcd5ead4e1c04dfa1, 0x36982c3d96d9ec47, 0x7341db869a7d531d, "If the enemy is within range, then so are you."},
	{0xd225e92e, 0xf458e174, 0xc2f8db8624fefc0e, 0xc2f8db8624fefc0e, 0x755a46f13a764857, 0xe7f5413a7f308fd6, 0x41d35389237b36e4, 0xc544a2d600ae8dfb, 0x619e9c54f1c9f3d4, 0xdcf6f01ae331847e, "The major problem is with sendmail.  -Mark Horton"},
	{0x0819a4e8, 0xae0b7a86, 0x5a0a6efd52e84e2a, 0x5a0a6efd52e84e2a, 0x3eb089865ab8b75e, 0x9fb5f8a3ed197d12, 0x9813f2f5d9b33cb3, 0xc59d72687a4b9b15, 0x2f9635cc9d58655a, 0xc219bcc95498f19f, "How can you write a big system without C++?  -Paul Glick"},
	{0xf585dfc4, 0x8e82e216, 0x786d7e1987023ca9, 0x786d7e1987023ca9, 0x54bf932850a761e1, 0x636c00055b1d3e3f, 0x273ef578b7c1056b, 0xa872b4052ea8c636, 0xa695b68b07c0d24c, 0x715add30687d7092, "He who has a shady past knows that nice guys finish last."},
	{0x7613810f, 0x3f6469cc, 0x5d14f96c18fe3d5e, 0x5d14f96c18fe3d5e, 0x7d87801f634222c4, 0xfa0e0a8b2a984142, 0x4813d464644d7658, 0x7367af890cc39d36, 0xc62a3e17cf5aff0b, 0x9f53c87726807e89, "Free! Free!/A trip/to Mars/for 900/empty jars/Burma Shave"},
	{0x1090d244, 0x84dd9748, 0xec8848fd3b266c10, 0xec8848fd3b266c10, 0xddce3cea9126d550, 0x8c085581e9f35148, 0xe09d3113753a527c, 0xec22d95954539646, 0xe09f6dfb52ddaedd, 0x3deedf6bf668881c, "His money is twice tainted: 'taint yours and 'taint mine."},
	{0x2cc30bb7, 0xa3672fa1, 0x2a578b80bb82147c, 0x2a578b80bb82147c, 0xf425894aebcce21a, 0x7a02d6d816a67950, 0xcccd07c8d398e7fe, 0x8f623e32dec3f91f, 0xe6b9e7bd64737d0b, 0xaebc6683db6a5ed3, "The days of the digital watch are numbered.  -Tom Stoppard"},
	{0xa5812ac8, 0x65c7e743, 0x55182f8859eca4ce, 0xb8d104d1135bbc60, 0xf0b5522f8062b30f, 0xb76f3e8c68eea2bc, 0x3e23a6e232671c25, 0x5ccc8ce07185764f, 0xfd24d1c7bef69416, 0xf22c727703f66aa4, "For every action there is an equal and opposite government program."},
	{0xd07de88f, 0x104a7d01, 0xabcdb319fcf2826c, 0x836f5ff0c2a7dfaa, 0x7b609bd1b51446df, 0xf45b1631b573f390, 0xd7ff0e47283d075c, 0xb7663fe8a39ee896, 0xdbdcf1735ff9b85c, 0x26a482705faec710, "You remind me of a TV show, but that's all right: I watch it anyway."},
	{0x2e18e880, 0xf7f48b6b, 0x1d85702503ac7eb4, 0x61751a90ec7d71bf, 0x4e58adfd902e4097, 0x141a2e0ff5e0e729, 0x1082910836b47d27, 0x2f3b24ebcfa58e41, 0x847e36b2945571f6, 0x4358a10fff41fcc0, "It's well we cannot hear the screams/That we create in others' dreams."},
	{0x1b8db5d0, 0x0159176d, 0xa2b8bf3032021993, 0xa58e3702193e4631, 0x50f9160881ddf6ff, 0xbbdfff333383198f, 0x1cdb19fd13eccaff, 0x28db4556ceb58337, 0xc8f51ab686f4ddaa, 0x74c85ad931043872, "Give me a rock, paper and scissors and I will move the world.  CCFestoon"},
	{0xcc3d0ff2, 0xc20708a8, 0x38aa3175b37f305c, 0xa104da6f2f575514, 0x88387579f38d77c9, 0xbfe4b3fa04bb6506, 0xabc0c75984ae62f2, 0x0ff3a7fb8f8419ae, 0xec996d7f77e0de82, 0x28a45862814611fe, "It's a tiny change to the code and not completely disgusting. - Bob Manchek"},
	{0xff16c9e6, 0x99735ad0, 0x7e85d7b050ed2967, 0xfe6aa49558b3cbe0, 0x249fa05ebf23cc74, 0x113915d477dc320b, 0x313bece7d506637d, 0x04e405567dabb986, 0x95d4a5700f5adee8, 0x71adea5e2345f3e9, "There is no reason for any individual to have a computer in their home. -Ken Olsen, 1977"},
	{0xe2053c2c, 0x510d0e3c, 0x5a05644eb66e435e, 0x96d012f9bccb3e, 0x7b88e6bee477c3e0, 0x07d20be82d1835c9, 0x89432ab05f44af82, 0xcdab02ccb904bcd1, 0x4edcb81c107d8234, 0xfc94e1b62a99beb3, "Even if I could be Shakespeare, I think I should still choose to be Faraday. - A. Huxley"},
	{0x11c493bb, 0xa36704a5, 0x98eff6958c5e91a, 0x8fe4429d157f60f5, 0x37d8c1e3d3f79073, 0xdf1461ed0c9c3fcb, 0x71dc055b46107f35, 0xff33cd37f0985850, 0xfaa96848001f77af, 0x1bd9b707750cf5ac, "The fugacity of a constituent in a mixture of gases at a given temperature is proportional to its mole fraction.  Lewis-Randall Rule"},
	{0xd402abf8, 0x656d5959, 0xc3f02c4ffd5d71e6, 0xed25cfc61b15bddd, 0x42b50de39d3f2225, 0x2e29484ace9fbcf0, 0x2012daecbb4ecf6c, 0xe75cabf03e64cbca, 0xefc0033cde0e7257, 0x8040fd1a39ea0f06, "Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum."},
}

func TestHash32(t *testing.T) {
	for _, tt := range testData {
		if h := Hash32([]byte(tt.in)); h != tt.oh32 {
			t.Errorf("Hash32(%q)=%#08x (len=%d), want %#08x", tt.in, h, len(tt.in), tt.oh32)
		}
	}
}

func TestFingerprint32(t *testing.T) {
	for _, tt := range testData {
		if h := Fingerprint32([]byte(tt.in)); h != tt.oh32 {
			t.Errorf("Hash32(%q)=%#08x (len=%d), want %#08x", tt.in, h, len(tt.in), tt.oh32)
		}
	}
}

func TestHash32WithSeed(t *testing.T) {
	for _, tt := range testData {
		if h := Hash32WithSeed([]byte(tt.in), 32); h != tt.oh32s {
			t.Errorf("Hash32WithSeed(%q)=%#08x (len=%d), want %#08x", tt.in, h, len(tt.in), tt.oh32s)
		}
	}
}

func TestHash64(t *testing.T) {
	for _, tt := range testData {
		if h := Hash64([]byte(tt.in)); h != tt.oh64 {
			t.Errorf("Hash64(%q)=%#08x (len=%d), want %#08x", tt.in, h, len(tt.in), tt.oh64)
		}
	}
}

func TestFingerprint64(t *testing.T) {
	for _, tt := range testData {
		if h := Fingerprint64([]byte(tt.in)); h != tt.oh64na {
			t.Errorf("Hash64(%q)=%#08x (len=%d), want %#08x", tt.in, h, len(tt.in), tt.oh64na)
		}
	}
}

func TestHash64WithSeed(t *testing.T) {
	for _, tt := range testData {
		if h := Hash64WithSeed([]byte(tt.in), 32); h != tt.oh64s {
			t.Errorf("Hash64WithSeed(%q)=%#08x (len=%d), want %#08x", tt.in, h, len(tt.in), tt.oh64s)
		}
	}
}

func TestHash64WithSeeds(t *testing.T) {
	for _, tt := range testData {
		if h := Hash64WithSeeds([]byte(tt.in), 32, 64); h != tt.oh64ss {
			t.Errorf("Hash64WithSeeds(%q)=%#08x (len=%d), want %#08x", tt.in, h, len(tt.in), tt.oh64ss)
		}
	}
}

func TestHash128(t *testing.T) {
	for _, tt := range testData {
		if lo, hi := Hash128([]byte(tt.in)); lo != tt.oh128lo || hi != tt.oh128hi {
			t.Errorf("Hash128(%q)=(%#016x, %#016x) (len=%d) want (%#016x, %#016x)", tt.in, lo, hi, len(tt.in), tt.oh128lo, tt.oh128hi)
		}
	}
}

func TestHash128WithSeed(t *testing.T) {
	for _, tt := range testData {
		if lo, hi := Hash128WithSeed([]byte(tt.in), 32, 64); lo != tt.oh128slo || hi != tt.oh128shi {
			t.Errorf("Hash128WithSeed(%q)=(%#016x, %#016x) (len=%d) want (%#016x, %#016x)", tt.in, lo, hi, len(tt.in), tt.oh128slo, tt.oh128shi)
		}
	}
}
