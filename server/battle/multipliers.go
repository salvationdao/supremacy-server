package battle

type MultiplierSystem struct {
	battle *Battle
}

func NewMultiplierSystem(btl *Battle) *MultiplierSystem {
	ms := &MultiplierSystem{
		btl,
	}
	ms.init()
	return ms
}

func (ms *MultiplierSystem) init() {

}
