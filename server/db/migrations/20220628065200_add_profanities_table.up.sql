CREATE TABLE profanities (
    id UUID PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
    phrase TEXT NOT NULL,
    UNIQUE (phrase)
);

INSERT INTO
    profanities (id, phrase)
VALUES
    (default, 'rape'),
    (default, '(ock'),
    (default, '[ock'),
    (default, '.f uc k'),
    (default, '@rse'),
    (default, '@rsehol'),
    (default, '@unt'),
    (default, '[unt'),
    (default, '< unt'),
    (default, '<***s'),
    (default, '<**t'),
    (default, '<**ts'),
    (default, '<**t''s'),
    (default, '<*nt'),
    (default, '<.unt'),
    (default, '<loth head'),
    (default, '<lothhead'),
    (default, '<nuts'),
    (default, '<o(k'),
    (default, '<o<&nbsp;k'),
    (default, '<o<ksu<ka'),
    (default, '<o<ksu<ker'),
    (default, '<oon'),
    (default, '<u&nbsp;nt'),
    (default, '<u&nbsp;nts'),
    (default, '<u*t'),
    (default, '<unt'),
    (default, '<unt''s'),
    (default, '<vnt'),
    (default, '<vnts'),
    (default, 'a$$hole'),
    (default, 'a$$hole$'),
    (default, 'a$$holes'),
    (default, 'a.rse'),
    (default, 'a+*hole'),
    (default, 'ar$ehole'),
    (default, 'ar$hole'),
    (default, 'ar$holes'),
    (default, 'ar5h0le'),
    (default, 'ar5h0les'),
    (default, 'ars3'),
    (default, 'arse hole'),
    (default, 'arseh0le'),
    (default, 'arseh0les'),
    (default, 'arseho'),
    (default, 'arsehol'),
    (default, 'arsehole'),
    (default, 'arseholes'),
    (default, 'arsewipe'),
    (default, 'arsh0le'),
    (default, 'arshole'),
    (default, 'arsholes'),
    (default, 'ashole'),
    (default, 'ass h0le'),
    (default, 'ass hole'),
    (default, 'assh0le'),
    (default, 'assh0les'),
    (default, 'asshole'),
    (default, 'assholes'),
    (default, 'b***ocks'),
    (default, 'b***ocks.'),
    (default, 'b*ll*cks'),
    (default, 'b.astard'),
    (default, 'b.ollocks'),
    (default, 'b.ugger'),
    (default, 'b@st@rd'),
    (default, 'b@st@rds'),
    (default, 'b00tha'),
    (default, 'b00thas'),
    (default, 'b0ll0cks'),
    (default, 'b0llocks'),
    (default, 'b1tch'),
    (default, 'b3llend'),
    (default, 'bastard'),
    (default, 'bastards'),
    (default, 'basterd'),
    (default, 'basyard'),
    (default, 'basyards'),
    (default, 'batty boy'),
    (default, 'batty&nbsp;boi'),
    (default, 'batty&nbsp;boy'),
    (default, 'beef curtains'),
    (default, 'belend'),
    (default, 'bell end'),
    (default, 'bell.end'),
    (default, 'bellend'),
    (default, 'bell-end'),
    (default, 'bin dippers'),
    (default, 'bin-dippers'),
    (default, 'bo****ks'),
    (default, 'bo11ocks'),
    (default, 'boabie sooking'),
    (default, 'boaby sooking'),
    (default, 'boll0cks'),
    (default, 'bollocks'),
    (default, 'bollox'),
    (default, 'bolocks'),
    (default, 'bolox'),
    (default, 'boner'),
    (default, 'boners'),
    (default, 'bootha'),
    (default, 'boothas'),
    (default, 'bugger'),
    (default, 'bukkake'),
    (default, 'bum bandit'),
    (default, 'bum hole'),
    (default, 'bumbandit'),
    (default, 'bum-bandit'),
    (default, 'bumh0l3'),
    (default, 'bumh0le'),
    (default, 'bumhol3'),
    (default, 'bumhole'),
    (default, 'c *nt!'),
    (default, 'c *nts!'),
    (default, 'c u n t'),
    (default, 'c u n t.'),
    (default, 'c#nt'),
    (default, 'c&nbsp;u&nbsp;n&nbsp;t'),
    (default, 'c* nt'),
    (default, 'c***s'),
    (default, 'c**k'),
    (default, 'c**t'),
    (default, 'c**ts'),
    (default, 'c**t''s'),
    (default, 'c*nt'),
    (default, 'c.u.n.t'),
    (default, 'c.unt'),
    (default, 'c.untyb.ollocks'),
    (default, 'c_u_n_t'),
    (default, 'c00n'),
    (default, 'c0ck'),
    (default, 'c0cksucka'),
    (default, 'c0cksucker'),
    (default, 'cahnt'),
    (default, 'cahnts'),
    (default, 'clit'),
    (default, 'clunge'),
    (default, 'cnut'),
    (default, 'cnuts'),
    (default, 'co(k'),
    (default, 'coc&nbsp;k'),
    (default, 'cock'),
    (default, 'cocksucka'),
    (default, 'cocksucker'),
    (default, 'cocksuckers'),
    (default, 'cocksuckers.'),
    (default, 'coon'),
    (default, 'cossor ali'),
    (default, 'cretin'),
    (default, 'cripple'),
    (default, 'critest'),
    (default, 'cu&nbsp;nt'),
    (default, 'cu&nbsp;nts'),
    (default, 'cu*t'),
    (default, 'cunt'),
    (default, 'c-u-n-t'),
    (default, 'cunting'),
    (default, 'cunts'),
    (default, 'cunt''s'),
    (default, 'cvnt'),
    (default, 'cvnts'),
    (default, 'd**khead'),
    (default, 'd1ck'),
    (default, 'd1ck!'),
    (default, 'd1ckh@ed'),
    (default, 'darkie'),
    (default, 'darky'),
    (default, 'dick'),
    (default, 'dick&nbsp;head'),
    (default, 'dickhead'),
    (default, 'dumbfuck'),
    (default, 'dumbfucker'),
    (default, 'dxxkhead'),
    (default, 'ethnics'),
    (default, 'f ck'),
    (default, 'f o a d'),
    (default, 'f off'),
    (default, 'f u c k'),
    (default, 'f u c ked'),
    (default, 'f uc k'),
    (default, 'f uc king'),
    (default, 'f uck'),
    (default, 'f###'),
    (default, 'f##k'),
    (default, 'f##king'),
    (default, 'f#cked'),
    (default, 'f$cks'),
    (default, 'f&nbsp;cked'),
    (default, 'f&nbsp;u&nbsp;c&nbsp;k'),
    (default, 'f&nbsp;uck'),
    (default, 'f&nbsp;ucker'),
    (default, 'f&nbsp;ucking'),
    (default, 'f()()k'),
    (default, 'f()()ker'),
    (default, 'f()()king'),
    (default, 'f*#kin'''),
    (default, 'f*&k'),
    (default, 'f*&k!ng'),
    (default, 'f***'),
    (default, 'f*****'),
    (default, 'f******'),
    (default, 'f*****g'),
    (default, 'f****d'),
    (default, 'f***ed'),
    (default, 'f***in'),
    (default, 'f***ing'),
    (default, 'f**k'),
    (default, 'f**ked'),
    (default, 'f**ker'),
    (default, 'f**kin'),
    (default, 'f**king'),
    (default, 'f**ks'),
    (default, 'f*c*'),
    (default, 'f*ck'),
    (default, 'f*ckin'),
    (default, 'f*ked'),
    (default, 'f*ker'),
    (default, 'f*ks'),
    (default, 'f*uck'),
    (default, 'f*uk'),
    (default, 'f.o.a.d'),
    (default, 'f.o.a.d.'),
    (default, 'f.u.c.k.'),
    (default, 'f.uck'),
    (default, 'f@@@in'),
    (default, 'f@@@ing'),
    (default, 'f@ck'),
    (default, 'f@g'),
    (default, 'f@gs'),
    (default, 'f[_]cker'),
    (default, 'f[_]cking'),
    (default, 'f^^k'),
    (default, 'f^^ked'),
    (default, 'f^^ker'),
    (default, 'f^^king'),
    (default, 'f^ck'),
    (default, 'f^cker'),
    (default, 'f^cking'),
    (default, 'f__kin'),
    (default, 'f__king'),
    (default, 'f<uk'),
    (default, 'f>>k'),
    (default, 'f00k'),
    (default, 'f00ked'),
    (default, 'f0oked'),
    (default, 'fack'),
    (default, 'fackin'),
    (default, 'facking'),
    (default, 'fag'),
    (default, 'fagg0t'),
    (default, 'faggits'),
    (default, 'faggot'),
    (default, 'fagits'),
    (default, 'fags'),
    (default, 'fanny'),
    (default, 'fc*king'),
    (default, 'fck'),
    (default, 'f''ck'),
    (default, 'fck&nbsp;ing'),
    (default, 'fck1ng'),
    (default, 'fckeud'),
    (default, 'fckin'),
    (default, 'fcking'),
    (default, 'fcks'),
    (default, 'fckw1t'),
    (default, 'fckwit'),
    (default, 'fcuk'),
    (default, 'fcuked'),
    (default, 'fcuker'),
    (default, 'fcukin'),
    (default, 'fcuking'),
    (default, 'fcuks'),
    (default, 'feck'),
    (default, 'feckin'),
    (default, 'fecking'),
    (default, 'f---ed'),
    (default, 'fekking'),
    (default, 'felch'),
    (default, 'felched'),
    (default, 'felching'),
    (default, 'feltch'),
    (default, 'feltcher'),
    (default, 'feltching'),
    (default, 'f-----g'),
    (default, 'f---ing'),
    (default, 'f--k'),
    (default, 'fkin'),
    (default, 'fking'),
    (default, 'flange'),
    (default, 'flucknuts'),
    (default, 'fo0ked'),
    (default, 'foad'),
    (default, 'f-o-a-d'),
    (default, 'fook'),
    (default, 'fookd'),
    (default, 'fooked'),
    (default, 'fooker'),
    (default, 'fookin'),
    (default, 'fookin'''),
    (default, 'fooking'),
    (default, 'frig'),
    (default, 'frigging'),
    (default, 'frigin'),
    (default, 'friging'),
    (default, 'fu <k'),
    (default, 'fu&kin'),
    (default, 'fu&king'),
    (default, 'fu&nbsp;ck'),
    (default, 'fu&nbsp;cked'),
    (default, 'fu&nbsp;cker'),
    (default, 'fu&nbsp;cking'),
    (default, 'fu(k'),
    (default, 'fu(ker'),
    (default, 'fu(kers'),
    (default, 'fu*k'),
    (default, 'fu.ck'),
    (default, 'fu@k'),
    (default, 'fu@ker'),
    (default, 'fu^k'),
    (default, 'fu^ker'),
    (default, 'fu^king'),
    (default, 'fu< kin'),
    (default, 'fu<k'),
    (default, 'f-u-<-k'),
    (default, 'fu<ked'),
    (default, 'fu<ker'),
    (default, 'fu<kin'),
    (default, 'fu<king'),
    (default, 'fu<kker'),
    (default, 'fu<kr'),
    (default, 'fu<ks'),
    (default, 'fuc&nbsp;ked'),
    (default, 'fuc&nbsp;ker'),
    (default, 'fuc&nbsp;king'),
    (default, 'fuck'),
    (default, 'f-uck'),
    (default, 'fúck'),
    (default, 'fúçk'),
    (default, 'fùck'),
    (default, 'fûck'),
    (default, 'fück'),
    (default, 'fuck&nbsp;ed'),
    (default, 'fuck&nbsp;ing'),
    (default, 'fucke&nbsp;d'),
    (default, 'fucked'),
    (default, 'fucker'),
    (default, 'fuckers'),
    (default, 'fucki&nbsp;ng'),
    (default, 'fuckin'),
    (default, 'fucking'),
    (default, 'fúcking'),
    (default, 'fuckinghell'),
    (default, 'fuckk'),
    (default, 'fucks'),
    (default, 'fuckup'),
    (default, 'fuckw1t'),
    (default, 'fuckwit'),
    (default, 'fuck-wit'),
    (default, 'fuckwits'),
    (default, 'fucw1t'),
    (default, 'fucwit'),
    (default, 'fudge p@cker'),
    (default, 'fudge packer'),
    (default, 'fudgep@cker'),
    (default, 'fudge-p@cker'),
    (default, 'fudgepacker'),
    (default, 'fudge-packer'),
    (default, 'fudgpacker'),
    (default, 'fuk'),
    (default, 'fukced'),
    (default, 'fuked'),
    (default, 'fuker'),
    (default, 'fukin'),
    (default, 'fuking'),
    (default, 'fukk'),
    (default, 'fukked'),
    (default, 'fukker'),
    (default, 'fukkin'),
    (default, 'fukking'),
    (default, 'fuks'),
    (default, 'fuukn'),
    (default, 'fvck'),
    (default, 'fvckup'),
    (default, 'fvck-up'),
    (default, 'fvckw1t'),
    (default, 'fvckwit'),
    (default, 'gang bang'),
    (default, 'gangbang'),
    (default, 'gang-bang'),
    (default, 'gash'),
    (default, 'gayhole'),
    (default, 'gimp'),
    (default, 'girlie-gardening'),
    (default, 'goris'),
    (default, 'gypo'),
    (default, 'gypos'),
    (default, 'gyppo'),
    (default, 'gyppos'),
    (default, 'h0m0'),
    (default, 'h0mo'),
    (default, 'homo'),
    (default, 'hvns'),
    (default, 'israelians'),
    (default, 'ities'),
    (default, 'jungle bunny'),
    (default, 'k**t'),
    (default, 'k@ffir'),
    (default, 'k@ffirs'),
    (default, 'k@fir'),
    (default, 'k@firs'),
    (default, 'kaf1r'),
    (default, 'kaff1r'),
    (default, 'kaffir'),
    (default, 'kaffirs'),
    (default, 'kafir'),
    (default, 'kafirs'),
    (default, 'kafr'),
    (default, 'kants'),
    (default, 'khunt'),
    (default, 'kiddie fiddler'),
    (default, 'kiddie fiddling'),
    (default, 'kiddie-fiddler'),
    (default, 'kiddie-fiddling'),
    (default, 'kiddy fiddler'),
    (default, 'kiddyfiddler'),
    (default, 'kiddy-fiddler'),
    (default, 'kike'),
    (default, 'kn0b'),
    (default, 'knob'),
    (default, 'knob&nbsp;head'),
    (default, 'knobber'),
    (default, 'knobhead'),
    (default, 'kraut'),
    (default, 'kuffar'),
    (default, 'kunt'),
    (default, 'kyke'),
    (default, 'l m f a o'),
    (default, 'l.m.f.a.o'),
    (default, 'l.m.f.a.o.'),
    (default, 'lmfa0'),
    (default, 'lmfao'),
    (default, 'm.inge'),
    (default, 'm.otherf.ucker'),
    (default, 'm1nge'),
    (default, 'minge'),
    (default, 'mof**ker'),
    (default, 'mof**kers'),
    (default, 'mofuccer'),
    (default, 'mofucker'),
    (default, 'mofuckers'),
    (default, 'mofucking'),
    (default, 'mofukcer'),
    (default, 'mohterf**ker'),
    (default, 'mohterf**kers'),
    (default, 'mohterf*kcer'),
    (default, 'mohterfuccer'),
    (default, 'mohterfuccers'),
    (default, 'mohterfuck'),
    (default, 'mohterfucker'),
    (default, 'mohterfuckers'),
    (default, 'mohterfucking'),
    (default, 'mohterfucks'),
    (default, 'mohterfuk'),
    (default, 'mohterfukcer'),
    (default, 'mohterfukcers'),
    (default, 'mohterfuking'),
    (default, 'mohterfuks'),
    (default, 'moterf**ker'),
    (default, 'moterfuccer'),
    (default, 'moterfuck'),
    (default, 'moterfucker'),
    (default, 'moterfuckers'),
    (default, 'moterfucking'),
    (default, 'moterfucks'),
    (default, 'mothaf**k'),
    (default, 'mothaf**ker'),
    (default, 'mothaf**kers'),
    (default, 'mothaf**king'),
    (default, 'mothaf**ks'),
    (default, 'mothafuccer'),
    (default, 'mothafuck'),
    (default, 'mothafucka'),
    (default, 'motha-fucka'),
    (default, 'mothafucker'),
    (default, 'mothafuckers'),
    (default, 'mothafucking'),
    (default, 'mothafucks'),
    (default, 'mother f---ers'),
    (default, 'motherf**ked'),
    (default, 'motherf**ker'),
    (default, 'motherf**kers'),
    (default, 'motherfuccer'),
    (default, 'motherfuccers'),
    (default, 'motherfuck'),
    (default, 'motherfucked'),
    (default, 'motherfucker'),
    (default, 'motherfuckers'),
    (default, 'motherfucking'),
    (default, 'motherfucks'),
    (default, 'motherfukkker'),
    (default, 'mthaf**ka'),
    (default, 'mthafucca'),
    (default, 'mthafuccas'),
    (default, 'mthafucka'),
    (default, 'mthafuckas'),
    (default, 'mthafukca'),
    (default, 'mthafukcas'),
    (default, 'muff'),
    (default, 'muff-diver'),
    (default, 'muff-diving'),
    (default, 'muffs'),
    (default, 'muth@fucker'),
    (default, 'muthaf**k'),
    (default, 'muthaf**ker'),
    (default, 'muthaf**kers'),
    (default, 'muthaf**king'),
    (default, 'muthaf**ks'),
    (default, 'muthafuccer'),
    (default, 'muthafuck'),
    (default, 'muthafuck@'),
    (default, 'muthafucka'),
    (default, 'muthafucker'),
    (default, 'muthafuckers'),
    (default, 'muthafucking'),
    (default, 'muthafucks'),
    (default, 'muthafukas'),
    (default, 'nig nog'),
    (default, 'niga'),
    (default, 'nigga'),
    (default, 'niggaz'),
    (default, 'nigger'),
    (default, 'niggers'),
    (default, 'nignog'),
    (default, 'nig-nog'),
    (default, 'nob&nbsp;head'),
    (default, 'nobhead'),
    (default, 'nonce'),
    (default, 'p**i'),
    (default, 'p*ki'),
    (default, 'p.iss-flaps'),
    (default, 'p@ki'),
    (default, 'p@kis'),
    (default, 'p00f'),
    (default, 'p00fs'),
    (default, 'p00fter'),
    (default, 'p00fters'),
    (default, 'p0of'),
    (default, 'paedo'),
    (default, 'paedophile'),
    (default, 'paedophiles'),
    (default, 'pak1'),
    (default, 'paki'),
    (default, 'pakis'),
    (default, 'peado'),
    (default, 'peadofile'),
    (default, 'peadofiles'),
    (default, 'peedo'),
    (default, 'peedofile'),
    (default, 'peedofiles'),
    (default, 'peedophile'),
    (default, 'peedophiles'),
    (default, 'peedos'),
    (default, 'pench0d'),
    (default, 'pench0ds'),
    (default, 'penchod'),
    (default, 'penchods'),
    (default, 'phanny'),
    (default, 'phanny.'),
    (default, 'pheck'),
    (default, 'phecking'),
    (default, 'phelching'),
    (default, 'pheque'),
    (default, 'phequing'),
    (default, 'phuck'),
    (default, 'phucker'),
    (default, 'phuckin'),
    (default, 'phucking'),
    (default, 'phucks'),
    (default, 'phuk'),
    (default, 'pikey'),
    (default, 'pillow biter'),
    (default, 'pillowbiter'),
    (default, 'pillow-biter'),
    (default, 'piss off'),
    (default, 'pissflaps'),
    (default, 'po0f'),
    (default, 'poff'),
    (default, 'ponce'),
    (default, 'poo stabber'),
    (default, 'poo stabbers'),
    (default, 'poof'),
    (default, 'poofs'),
    (default, 'poofter'),
    (default, 'pr!ck'),
    (default, 'pr!ck.'),
    (default, 'pr1ck'),
    (default, 'pr1ck!'),
    (default, 'pr1cks'),
    (default, 'pr1cks!'),
    (default, 'prick'),
    (default, 'prik'),
    (default, 'pu$$y'),
    (default, 'pussy'),
    (default, 'queer'),
    (default, 'raghead'),
    (default, 'ragheads'),
    (default, 'ret@rd'),
    (default, 'retard'),
    (default, 'rim job'),
    (default, 'rimming'),
    (default, 's.hit'),
    (default, 's1ut'),
    (default, 'sc u m!'),
    (default, 'sc um'),
    (default, 'sh hit'),
    (default, 'sh!t'),
    (default, 'sh!te'),
    (default, 'sh!tes'),
    (default, 'sh1t'),
    (default, 'sh1te'),
    (default, 'shirtlifters'),
    (default, 'shit stabber'),
    (default, 'shit stabbers'),
    (default, 'shitstabber'),
    (default, 'shitstabbers'),
    (default, 'spastic'),
    (default, 'spaz'),
    (default, 'spaz.'),
    (default, 'spic'),
    (default, 'spit roasting'),
    (default, 'spitroast'),
    (default, 'spit-roast'),
    (default, 'spit-roasting'),
    (default, 'spunk'),
    (default, 'spunking'),
    (default, 'ß0ll0çk5'),
    (default, 't w a t'),
    (default, 't wat'),
    (default, 't&nbsp;w&nbsp;a&nbsp;t'),
    (default, 't&nbsp;w&nbsp;a&nbsp;t&nbsp;s'),
    (default, 't*w*a*t'),
    (default, 't.wat'),
    (default, 't0$$er'),
    (default, 't0sser'),
    (default, 't0ssers'),
    (default, 'teabagging'),
    (default, 'tea-bagging'),
    (default, 'to55er'),
    (default, 'to55ers'),
    (default, 'tosser,'),
    (default, 'tossers'),
    (default, 'tossurs'),
    (default, 'towel head'),
    (default, 'towelhead'),
    (default, 'turd'),
    (default, 'tvvat'),
    (default, 'tvvats'),
    (default, 'tw at'),
    (default, 'tw&nbsp;at'),
    (default, 'tw@'),
    (default, 'tw@t'),
    (default, 'tw@ts'),
    (default, 'tw_t'),
    (default, 'tw4t'),
    (default, 'twa t'),
    (default, 'twat'),
    (default, 'twats'),
    (default, 'twatt'),
    (default, 'twattish'),
    (default, 'twunt'),
    (default, 'twunts'),
    (default, 'up the gary'),
    (default, 'w anker'),
    (default, 'w ankers'),
    (default, 'w#nker'),
    (default, 'w#nkers'),
    (default, 'w***er'),
    (default, 'w*nkers, 0'),
    (default, 'w.ank'),
    (default, 'w@nk'),
    (default, 'w@nker'),
    (default, 'w@nkers'),
    (default, 'w@nks'),
    (default, 'w0g'),
    (default, 'w0gs'),
    (default, 'w4nker!'),
    (default, 'w4nkers!'),
    (default, 'wa nker'),
    (default, 'wan k er'),
    (default, 'wan k ers'),
    (default, 'wan ker'),
    (default, 'wank'),
    (default, 'wanka'),
    (default, 'wanke r'),
    (default, 'wanked'),
    (default, 'wanker'),
    (default, 'wankers'),
    (default, 'wanking'),
    (default, 'wanks'),
    (default, 'wank''s'),
    (default, 'wet spam'),
    (default, 'whanker'),
    (default, 'whankers'),
    (default, 'wog'),
    (default, 'wop'),
    (default, 'xrse'),
    (default, 'xrseh'),
    (default, 'xrsehol'),
    (default, 'xrsehole'),
    (default, 'xxxhole'),
    (default, 'y!ddo!'),
    (default, 'y!ddo!!'),
    (default, 'y*d'),
    (default, 'yid'),
    (default, 'yido'),
    (default, 'zachariah bishop'),
    (
        default,
        'arse:arsenal,arsenals,arsenate,arsenates,arsenic,arsenical,arsenicalism,arsenics,arsenide,arsenides,arsenious,arsenite,arsenites,arsenium,arseniuret,arseniuretted,arsenopyrite,arsenotherapies,arsenotherapy,arsenous,arsenoxide,carse,catharses,charset,coarse,coarsely,coarsen,coarsened,coarseness,coarsening,coarsens,coarser,coarsest,farsee,farseeing,farseing,hearse,hearsed,hearses,hoarse,hoarsely,hoarsen,hoarsened,hoarseness,hoarsening,hoarsens,hoarser,hoarsest,katharses,larsen,marse,marseillaise,marseille,marseilles,marses,parse,parsec,parsecs,parsed,parsee,parser,parsers,parses,psychocatharses,rehearse,rehearsed,rehearser,rehearsers,rehearses,sarsen,sarsenet,sarsent,sparse,sparsely,sparseness,sparser,sparsest,tarsectomies,tarsectomy,unrehearse,unrehearsed'
    ),
    (default, 'asswipe'),
    (default, 'blowjob'),
    (default, 'blow-job'),
    (default, 'bollock'),
    (default, 'boner:deboner'),
    (default, 'bonk:bonkers'),
    (default, 'bullshit'),
    (
        default,
        'bugger:debugger,debuggers,humbugger,humbuggers,jitterbugger'
    ),
    (default, 'bunghole'),
    (default, 'candy-ass'),
    (default, 'chuffnuts'),
    (
        default,
        'clit:anaclitic,asynclitism,choroidocyclitis,clitellum,clitia,clitic,clition,clitoral,clitoric,clitoridean,clitoridectomies,clitoridectomy,clitoridis,clitoriditis,clitoris,clitorises,clitoritis,cyclitis,cyclitol,enclitic,enclitically,glaucomatocyclitic,heraclitus,heteroclite,iridocyclitis,polyclitus,proclitic,synclitic'
    ),
    (default, 'clitty'),
    (
        default,
        'cock:alcock,babcock,ballcock,bawcock,bibcock,billycock,blackcock,cock-a-doodle-doo,cock-a-hoop,cock-a-leekie,cock-eyed,cock-of-the-rock,cock-sparrow,cockade,cockaded,cockades,cockaigne,cockalorum,cockamamie,cockarouse,cockateel,cockateels,cockatiel,cockatoo,cockatoos,cockatrice,cockatrices,cockayne,cockbilled,cockboat,cockchafer,cockcroft,cockcrow,cockcrows,cocked,cocker,cockerel,cockerels,cockers,cockeye,cockeyed,cockeyes,cockfight,cockfighting,cockfights,cockhorse,cockhorses,cockier,cockiest,cockily,cockiness,cocking,cockish,cockle,cockleboat,cocklebur,cockled,cockles,cockleshell,cockleshells,cockloft,cockness,cockney,cockneyfy,cockneyism,cockneys,cockoldries,cockpit,cockpits,cockroach,cockroaches,cocks,cockscomb,cockscombs,cocksfoot,cockshy,cockspur,cockspurs,cocksure,cockswain,cocktail,cocktailed,cocktails,cockup,cockups,cocky,corncockle,gamecock,gamecocking,gamecocks,gorcock,half-cock,half-cocked,hancock,haycock,haycocks,hitchcock,leacock,moorcock,peacock,peacocked,peacockier,peacocking,peacocks,petcock,petcocks,pinchcock,poppycock,seacock,shuttlecock,shuttlecocks,spatchcock,spatchcocked,spatchcocking,spitchcock,stopcock,stopcocks,storm-cock,turncock,uncock,weathercock,weathercocks,woodcock,woodcocking,woodcocks'
    ),
    (default, 'cojones'),
    (
        default,
        'coon:barracoon,cocoon,cocooned,cocooning,cocoons,cooncan,coonhound,coonhounds,coons,coonskin,coonskins,coontie,laocoon,puccoon,raccoon,raccoons,racoon,racoons,tycoon,tycoons'
    ),
    (
        default,
        'cum:accumbent,accumulable,accumulate,accumulated,accumulates,accumulating,accumulation,accumulations,accumulative,accumulatively,accumulativeness,accumulator,accumulators,acumen,acumens,acuminata,acuminate,acuminated,acuminating,acumination,altocumulus,cacuminal,caecum,canonicum,capsicum,capsicums,caroticum,cecum,circum,circum-,circumambience,circumambient,circumambulate,circumambulated,circumambulates,circumambulating,circumambulation,circumambulations,circumanal,circumarticular,circumbendibus,circumcircle,circumcise,circumcised,circumcises,circumcising,circumcision,circumcisions,circumcorneal,circumduction,circumductions,circumference,circumferences,circumferential,circumflex,circumflexes,circumfluent,circumfluous,circumforaneous,circumfuse,circumfusion,circumgyration,circuminsular,circumjacence,circumjacent,circumlental,circumlocution,circumlocutions,circumlocutory,circumlunar,circumnavigate,circumnavigated,circumnavigates,circumnavigating,circumnavigation,circumnavigations,circumnavigator,circumnutate,circumoral,circumpolar,circumrotation,circumscissile,circumscribe,circumscribed,circumscribes,circumscribing,circumscription,circumscriptions,circumsolar,circumspect,circumspection,circumspectly,circumspectness,circumsphere,circumstance,circumstanced,circumstances,circumstantial,circumstantiality,circumstantially,circumstantiate,circumstantiated,circumstantiates,circumstantiating,circumstantiation,circumstantiations,circumstantibus,circumvallate,circumvallation,circumvascular,circumvent,circumventable,circumvented,circumventing,circumvention,circumventions,circumvents,circumvolution,circumvolve,cirrocumulus,colchicum,cucumber,cucumbers,cumae,cuman,cumarin,cumber,cumbered,cumberer,cumberers,cumbering,cumberland,cumberlandrian,cumbernauld,cumbers,cumbersome,cumbersomeness,cumbrance,cumbria,cumbrous,cumbrously,cumbrousness,cumin,cumins,cummerbund,cummerbunds,cummers,cummin,cummings,cummins,cumquat,cumquats,cumshaw,cumshaws,cumulate,cumulated,cumulates,cumulating,cumulation,cumulative,cumulatively,cumulet,cumuli,cumuliform,cumulonimbus,cumulostratus,cumulous,cumulus,curcuma,decumbence,decumbency,decumbent,disencumber,disencumbered,disencumbering,disencumbers,document,documentable,documental,documentaries,documentarily,documentary,documentation,documented,documenter,documenters,documenting,documents,doronicum,eboracum,ecclesiasticum,ecumenic,ecumenical,ecumenicalism,ecumenically,ecumenicism,ecumenicity,ecumenism,ecumenist,elasticum,encumber,encumbered,encumbering,encumbers,encumbrance,encumbrancer,encumbrances,fractocumulus,guaiacum,hypericum,ileocecum,illyricum,incumbencies,incumbency,incumbent,incumbently,incumbents,incumber,incumbered,incumbering,incumbers,incumbrance,incumbrancer,locum,locumtenencies,locumtenency,macumba,mecum,mecums,mesocecum,modicum,modicums,molluscum,nocumetum,noncumulative,noricum,oecumenical,procumbent,publicum,pyogenicum,recumbencies,recumbency,recumbent,rusticum,scum,scumbag,scumble,scumbled,scumbling,scummed,scummers,scummier,scummiest,scumming,scummy,scums,slocum,stratocumuli,stratocumulus,succumb,succumbed,succumber,succumbers,succumbing,succumbs,superincumbent,talcum,talcums,taraxacum,tecum,tecumseh,tillicum,triticum,tucum,uncircumcise,uncircumcised,uncircumcision,uncircumstantial,uncircumstantialy,uncumber,uncumbered,undocument,undocumented,unencumber,unencumbered,unincumber,unincumbered,vademecum,viaticum,viaticums'
    ),
    (default, 'cunny'),
    (default, 'cunt:scunthorpe'),
    (
        default,
        'crap:crapaud,crape,craped,crapes,craping,crapper,crappers,crappie,crappieness,crappier,crappies,crappiest,crappiness,crapping,crappy,craps,crapshooter,crapshooters,crapulence,crapulent,crapulous,scrap,scrapbook,scrapbooks,scrape,scraped,scraper,scraperboard,scrapers,scrapes,scrapheap,scrapie,scraping,scrapings,scrappage,scrapped,scrapper,scrappers,scrappier,scrappiest,scrappily,scrappiness,scrapping,scrapple,scrapples,scrappy,scraps,skyscrape,skyscraper,skyscrapers,skyscraping'
    ),
    (
        default,
        'dago:dagoba,dagobas,dagoes,dagon,pedagog,pedagogic,pedagogical,pedagogically,pedagogics,pedagogies,pedagogs,pedagogue,pedagogues,pedagogy,solidago,solidagos'
    ),
    (default, 'dipstick'),
    (
        default,
        'dong:ding-dong,dingdong,dingdonged,dingdongs,donga,dongola,dongs,quandong'
    ),
    (default, 'dork:dorking'),
    (default, 'feak'),
    (default, 'fecal:fecalith,fecaloid'),
    (default, 'fellate'),
    (default, 'furbox'),
    (default, 'furburger'),
    (default, 'gayboy'),
    (default, 'ginch'),
    (default, 'gnikcuf'),
    (default, 'hardon'),
    (default, 'honkers'),
    (default, 'hussy'),
    (default, 'kcid'),
    (default, 'kcuf'),
    (default, 'lactoids'),
    (
        default,
        'lesbo:blesbok,middlesboro,middlesborough'
    ),
    (default, 'lesbyterian'),
    (default, 'lezzie'),
    (default, 'lezzo'),
    (default, 'man-root'),
    (default, 'nestlecock'),
    (
        default,
        'nigger:snigger,sniggered,sniggering,sniggeringly,sniggers'
    ),
    (
        default,
        'nympho:nympholepsies,nympholepsy,nympholept,nympholeptic,nymphomania,nymphomaniac,nymphomaniacal,nymphomaniacs'
    ),
    (default, 'onanism'),
    (
        default,
        'piss:inspissate,inspissated,inspissating,inspissation,inspissator,nipissing,pissant,pissants,pissarro,pissoir,pissoirs,spissitude'
    ),
    (default, 'pissoff'),
    (
        default,
        'prick:pinprick,pinpricked,pinpricks,pricked,pricket,prickier,prickiest,pricking,prickle,prickled,prickles,pricklier,prickliest,prickliness,prickling,prickly,pricks,pricky'
    ),
    (default, 'pussies'),
    (
        default,
        'pussy:pussycat,pussycats,pussyfoot,pussyfooted,pussyfooting,pussyfoots'
    ),
    (default, 'pusy'),
    (
        default,
        'puta:amputate,amputated,amputates,amputating,amputation,amputations,amputator,computability,computable,computation,computational,computations,computative,deputation,deputational,deputations,deputative,disputability,disputable,disputably,disputant,disputants,disputation,disputations,disputatious,disputatiously,disputatiousness,disreputability,disreputable,disreputably,imputable,imputation,imputations,imputative,incomputable,incomputably,indisputable,indisputableness,indisputably,laputa,laputan,miscomputation,nonimputability,nonimputable,putamen,putamina,putative,putatively,rajputana,reamputation,reputability,reputable,reputableness,reputably,reputation,reputations,sputa,supputation,undisputable'
    ),
    (default, 'queef'),
    (default, 'queve'),
    (default, 'quim:equimolecular,esquimau,quimper'),
    (default, 'quimsteak'),
    (default, 'qveer'),
    (default, 'rimadonna:primadonna,primadonnas'),
    (
        default,
        'rimming:brimming,primming,trimming,trimmings,untrimming'
    ),
    (
        default,
        'rootle:rootless,rootlessness,rootlet,rootlets'
    ),
    (default, 'sappho'),
    (default, 'scumbag'),
    (default, 'scumber'),
    (default, 'sexpot'),
    (default, 'shag:shagbark,shagbarks,shagreen'),
    (default, 'shagbucket'),
    (default, 'shagstress'),
    (default, 'shirtlifter'),
    (
        default,
        'shit:brushite,cushitic,peshitta,shittah,shittim,yamashita'
    ),
    (default, 'snarf'),
    (default, 'sodomite'),
    (default, 'sodomy'),
    (
        default,
        'spic:allspice,allspices,aspic,aspics,auspicate,auspice,auspices,auspicial,auspicious,auspiciously,auspiciousness,conspicuity,conspicuous,conspicuously,conspicuousness,despicable,despicably,extispicious,haruspical,hospice,hospices,imperspicuity,inauspicious,inauspiciously,inauspiciousness,inconspicuous,inconspicuously,inconspicuousness,mispickel,oversuspicious,perspicacious,perspicaciously,perspicaciousness,perspicacity,perspicuity,perspicuous,perspicuously,perspicuousness,spica,spicaes,spicas,spicate,spiccato,spice,spiceberry,spicebush,spiced,spicer,spicers,spicery,spices,spicey,spicier,spiciest,spicily,spiciness,spicing,spick,spick-and-span,spics,spicula,spicular,spiculate,spiculated,spiculation,spicule,spicules,spiculum,spicy,suspicion,suspicions,suspicious,suspiciously,suspiciousness,transpicuous,unauspicious,unconspicuous,unsuspicious,unsuspiciously'
    ),
    (default, 'strollop'),
    (default, 'suckster'),
    (default, 'titties'),
    (default, 'tnuc'),
    (default, 'toggaf'),
    (default, 'tosser'),
    (default, 'tribadist'),
    (
        default,
        'turd:saturday,saturdayish,saturdays,sturdier,sturdies,sturdiest,sturdily,sturdiness,sturdy,turdiform,turdine'
    ),
    (default, 'twank'),
    (
        default,
        'twat:atwater,cutwater,derwentwater,heartwater,meltwater,saltwater,twattle,witwatersrand,wristwatch,wristwatches'
    ),
    (
        default,
        'wank:swank,swanked,swanker,swankers,swankest,swankier,swankiest,swankily,swanking,swanks,swanky,twankay'
    ),
    (default, 'whore'),
    (default, 'wiseass'),
    (default, 'wizzer'),
    (
        default,
        'wog:golliwog,golliwogs,hornswoggle,hornswoggled,hornswoggling,polliwog,polliwogs,pollywog,pollywogs,woggle'
    );