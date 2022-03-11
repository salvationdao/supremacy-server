-- insert NeoTokyo Map
INSERT INTO game_maps (name, max_spawns, image_url, width, height, cells_x, cells_y, top_pixels, left_pixels, scale, disabled_cells)
VALUES ('NeoTokyo',18,'https://ninjasoftware-static-media.s3.ap-southeast-2.amazonaws.com/supremacy/maps/neo_tokyo.jpg',1131,1191,30,32,-35000,-31000,0.0265,
        ARRAY[0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32, 33,
                   34, 35, 36, 37, 38, 39, 40, 43, 44, 45, 46, 47, 48, 49, 50, 51, 52, 53, 54, 55, 56, 57, 58, 59, 60, 61, 62, 63, 64, 65, 66, 67,
                   68, 69, 70, 74, 75, 76, 77, 78, 79, 80, 81, 82, 83, 84, 85, 86, 87, 88, 89, 90, 91, 92, 93, 94, 95, 96, 97, 98, 99, 100,
                   104, 105, 106, 107, 108, 109, 110, 111, 112, 113, 114, 115, 116, 117, 118, 119, 120, 121, 122, 123, 124, 125, 126, 127, 128, 129, 130, 135, 160,
                   169, 170, 171, 172, 173, 174, 175, 176, 177, 178, 179, 180, 181, 182, 183, 184, 185, 186, 187, 188, 189, 190, 199, 200, 201, 208, 209, 210, 211, 212, 213,
                   214, 215, 216, 217, 218, 219, 220, 229, 230, 231, 238, 239, 240, 241, 242, 243, 244, 245, 246, 247, 248, 249, 250, 259, 260, 261, 268, 269, 270, 271,
                   272, 273, 274, 275, 276, 277, 278, 279, 280, 298, 299, 300, 301, 302, 303, 304, 305, 306, 307, 308, 309, 310, 328, 329, 330, 331, 332, 333, 334, 335,
                   335, 336, 337, 338, 339, 340, 358, 359, 360, 361, 362, 363, 364, 365, 366, 367, 368, 369, 370, 388, 389, 390, 391, 392, 393, 394, 395, 396, 397,
                   398, 399, 400, 418, 419, 420, 421, 422, 423, 424, 425, 426, 427, 428, 429, 430, 448, 449, 450, 451, 452, 453, 454, 455, 456, 457, 458, 459, 460, 478, 479,
                   480, 481, 482, 483, 484, 485, 486, 487, 488, 489, 490, 508, 509, 510, 511, 512, 513, 514, 515, 516, 517, 518, 519, 520, 538, 539, 540, 541, 542, 543,
                   544, 545, 546, 547, 548, 549, 550, 568, 569, 570, 571, 572, 573, 574, 575, 576, 577, 578, 579, 580, 598, 599, 600, 601, 602, 603, 604, 605, 606, 607,
                   608, 609, 610, 628, 629, 630, 631, 632, 633, 634, 635, 636, 637, 638, 639, 640, 658, 659, 660, 661, 688, 689, 690, 691, 718, 719, 720, 721, 722,
                   748, 749, 750, 751, 752, 753, 778, 779, 780, 781, 782, 783, 784, 808, 809, 810, 811, 812, 813, 814, 815, 838, 839, 840, 841, 842, 843, 844, 845,
                   846, 868, 869, 870, 871, 872, 873, 874, 875, 876, 877, 898, 899, 900, 901, 902, 903, 904, 905, 906, 907, 908, 909, 910, 911, 912, 913, 914, 915, 916,
                   917, 918, 919, 920, 921, 922, 923, 924, 925, 926, 927, 928, 929, 930, 931, 932, 933, 934, 935, 936, 937, 938, 939, 940, 941, 942, 943, 944, 945, 946,
                   947, 948, 949, 950, 951, 952, 953, 954, 955, 956, 957, 958, 959]);

INSERT INTO game_maps (name, max_spawns, image_url, width, height, cells_x, cells_y, top_pixels, left_pixels, scale, disabled_cells)
VALUES ('UrbanBuildings', 32, 'https://ninjasoftware-static-media.s3.ap-southeast-2.amazonaws.com/supremacy/maps/urban_city.jpg', 889, 894, 45, 45, -76000, -118000, 0.00987777777777777777777777777778, ARRAY[
    0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32, 33,
    34, 35, 36, 37, 38, 39, 40, 41, 42, 43, 44, 45, 46, 47, 48, 49, 50, 51, 52, 53, 54, 55, 56, 57, 58, 59, 60, 61, 62, 63, 64, 65, 66, 67,
    68, 69, 70, 71, 72, 73, 74, 75, 76, 77, 78, 79, 80, 81, 82, 83, 84, 85, 86, 87, 88, 89, 90, 91, 92, 93, 94, 95, 96, 97, 98, 99, 100, 101,
    102, 103, 104, 105, 106, 107, 108, 109, 110, 111, 112, 113, 114, 115, 116, 117, 118, 119, 120, 121, 122, 123, 124, 125, 126, 127, 128, 129,
    130, 131, 132, 133, 134, 135, 136, 137,
    180, 181, 182, 225, 226, 227, 270, 271, 272, 315, 316, 317, 360, 361, 362, 405, 406, 407, 450, 451, 452, 495, 496, 497, 540, 541, 542, 585, 586, 587,
    630, 631, 632, 675, 676, 677, 720, 721, 722, 765, 766, 767, 810, 811, 812, 855, 856, 857, 900, 901, 902, 945, 946, 947, 990, 991, 992, 1035, 1036, 1037,
    1037, 1080, 1081, 1082, 1125, 1126, 1127, 1170, 1171, 1172, 1215, 1216, 1217, 1260, 1261, 1262, 1305, 1306, 1307, 1350, 1351, 1352, 1395, 1396, 1397,
    1440, 1441, 1442, 1485, 1486, 1487, 1530, 1531, 1532, 1575, 1576, 1577, 1620, 1621, 1622, 1665, 1666, 1667, 1710, 1711, 1712, 1755, 1756, 1757, 1800,
    1801, 1802, 1845, 1846, 1847, 1890, 1891, 1892, 1893, 1894, 1895, 1896, 1897, 1898, 1899, 1900, 1901, 1902, 1903, 1904, 1905, 1906, 1907, 1908, 1909, 1910,
    1911, 1912, 1913, 1914, 1915, 1916, 1917, 1918, 1919, 1920, 1921, 1922, 1923, 1924, 1925, 1926, 1927, 1928, 1929, 1930, 1931, 1932, 1933, 1934, 1935, 1936,
    1937, 1938, 1939, 1940, 1941, 1942, 1943, 1944, 1945, 1946, 1947, 1948, 1949, 1950, 1951, 1952, 1953, 1954, 1955, 1956, 1957, 1958, 1959, 1960, 1961,
    1962, 1963, 1964, 1965, 1966, 1967, 1968, 1969, 1970, 1971, 1972, 1973, 1974, 1975, 1976, 1977, 1978, 1979, 1980, 1981, 1982, 1983, 1984, 1985, 1986,
    1987, 1988, 1989, 1990, 1991, 1992, 1993, 1994, 1995, 1996, 1997, 1998, 1999, 2000, 2001, 2002, 2003, 2004, 2005, 2006, 2007, 2008, 2009, 2010, 2011,
    2012, 2013, 2014, 2015, 2016, 2017, 2018, 2019, 2020, 2021, 2022, 2023, 2024,
    177, 178, 179, 222, 223, 224, 267, 268, 269, 312, 313, 314, 357, 358, 359, 402, 403, 404, 447, 448, 449, 492, 493, 494, 537, 538, 539, 582, 583, 584,
    627, 628, 629, 672, 673, 674, 717, 718, 719, 762, 763, 764, 807, 808, 809, 852, 853, 854, 897, 898, 899, 942, 943, 944, 987, 988, 989, 1032, 1033, 1034,
    1077, 1078, 1079, 1122, 1123, 1124, 1167, 1168, 1169, 1212, 1213, 1214, 1257, 1258, 1259, 1302, 1303, 1304, 1347, 1348, 1349, 1437, 1438, 1439, 1482,
    1483, 1484, 1527, 1528, 1529, 1572, 1573, 1574, 1617, 1618, 1619, 1662, 1663, 1664, 1707, 1708, 1709, 1752, 1753, 1754, 1797, 1798, 1799, 1842, 1843,
    1844, 1887, 1888, 1889]);