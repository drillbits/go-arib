package xcs

var (
	_ GraphicSet = KanjiSet
	_ GraphicSet = AlphanumericSet
	_ GraphicSet = KatakanaSet
	_ GraphicSet = HiraganaSet
	_ GraphicSet = AdditionalSymbols
	_ GraphicSet = AdditionalKanjiCharacters
)
