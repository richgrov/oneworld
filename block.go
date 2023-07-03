package oneworld

import "github.com/richgrov/oneworld/blocks"

type Block struct {
	ty   blocks.BlockType
	data byte
}

func (block *Block) Type() blocks.BlockType {
	return block.ty
}

func (block *Block) Data() byte {
	return block.data
}
