package api

import (
	"context"

	"github.com/OnsagerHe/geoip-detector/pkg/utils"
	pb "github.com/OnsagerHe/geoip-detector/proto/gen"
)

func (r *Frontend) PutEndpoint(ctx context.Context, req *pb.PutEndpointRequest) (*[]pb.PutEndpointResponse, error) {
	r.Retriever.Process.Logger.Debug("call CheckStatus functions!")

	if req.GetLoop() != 0 {
		r.Retriever.Utils.Loop = uint8(req.Loop)
	}
	r.Retriever.Process.Resource = utils.EndpointMetadata{Endpoint: req.Endpoint}

	r.Retriever.Process.Logger.Debug("value for endpoint and loop:" + req.Endpoint)
	res, err := r.Retriever.CheckEndpoint()
	if err != nil {
		return res, err
	}
	return res, nil
}
