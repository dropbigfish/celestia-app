package ante

import (
	"cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/celestiaorg/go-square/v2/share"

	"github.com/celestiaorg/celestia-app/v4/pkg/appconsts"
	blobtypes "github.com/celestiaorg/celestia-app/v4/x/blob/types"
)

// MaxTotalBlobSizeDecorator helps to prevent a PFB from being included in a
// block but not fitting in a data square.
type MaxTotalBlobSizeDecorator struct {
	k BlobKeeper
}

func NewMaxTotalBlobSizeDecorator(k BlobKeeper) MaxTotalBlobSizeDecorator {
	return MaxTotalBlobSizeDecorator{k}
}

// AnteHandle implements the Cosmos SDK AnteHandler function signature. It
// returns an error if tx contains a MsgPayForBlobs where the total blob size is
// greater than the max total blob size.
func (d MaxTotalBlobSizeDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (sdk.Context, error) {
	if !ctx.IsCheckTx() {
		return next(ctx, tx, simulate)
	}

	maxSize := d.maxTotalBlobSize(ctx)
	for _, m := range tx.GetMsgs() {
		if pfb, ok := m.(*blobtypes.MsgPayForBlobs); ok {
			if total := getTotal(pfb.BlobSizes); total > maxSize {
				return ctx, errors.Wrapf(blobtypes.ErrTotalBlobSizeTooLarge, "total blob size %d exceeds max %d", total, maxSize)
			}
		}
	}

	return next(ctx, tx, simulate)
}

// maxTotalBlobSize returns the max number of bytes available for blobs in a
// data square based on the max square size. Note it is possible that txs with a
// total blob size less than this max still fail to be included in a block due
// to overhead from the PFB tx and/or padding shares.
func (d MaxTotalBlobSizeDecorator) maxTotalBlobSize(ctx sdk.Context) int {
	squareSize := d.getMaxSquareSize(ctx)
	totalShares := squareSize * squareSize
	// The PFB tx share must occupy at least one share so the # of blob shares
	// is at least one less than totalShares.
	blobShares := totalShares - 1
	return share.AvailableBytesFromSparseShares(blobShares)
}

// getMaxSquareSize returns the max effective square size.
func (d MaxTotalBlobSizeDecorator) getMaxSquareSize(ctx sdk.Context) int {
	govMax := d.k.GetParams(ctx).GovMaxSquareSize
	hardMax := appconsts.SquareSizeUpperBound
	return min(int(govMax), hardMax)
}

// getTotal returns the sum of the given sizes.
func getTotal(sizes []uint32) (sum int) {
	for _, size := range sizes {
		sum += int(size)
	}
	return sum
}
