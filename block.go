package oneworld

import "github.com/richgrov/oneworld/blocks"

type Block struct {
	ty   blocks.BlockType
	data byte
}

func NewBlock(ty blocks.BlockType, data byte) Block {
	return Block{ty, data}
}

func (block *Block) Type() blocks.BlockType {
	return block.ty
}

func (block *Block) Data() byte {
	return block.data
}
