package main

import (
	"net/http"

	pb "github.com/lightstaff/grpc_test/protobuf"

	"io"

	"github.com/labstack/echo"
	netCtx "golang.org/x/net/context"
)

// Heloと返すだけ
func GetHello(c echo.Context) error {
	sc, ok := c.(*ServiceContext)
	if !ok {
		return echo.NewHTTPError(http.StatusBadRequest, "コンテキストが取得できません")
	}

	rep, err := sc.ServiceClient.GetHello(netCtx.Background(), &pb.Empty{})
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"reply": rep.Result,
	})
}

// stream経由で受けた文字列を大文字化して返すサービスを呼び出してやりとり
func UpperCharacters(c echo.Context) error {
	sc, ok := c.(*ServiceContext)
	if !ok {
		return echo.NewHTTPError(http.StatusBadRequest, "コンテキストが取得できません")
	}

	type bodyModel struct {
		Messages []string `json:"messages"`
	}

	var m bodyModel
	if err := c.Bind(&m); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	stream, err := sc.ServiceClient.UpperCharacters(netCtx.Background())
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	errC := make(chan error)
	resultC := make(chan *pb.ReplyModel)
	doneC := make(chan struct{})
	go func() {
		defer func() {
			close(errC)
			close(resultC)
			close(doneC)
		}()

		for {
			res, err := stream.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				errC <- err
				return
			}
			resultC <- res
		}
	}()

	for _, message := range m.Messages {
		if err := stream.Send(&pb.ReqModel{
			Message: message,
		}); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}
	}

	if err := stream.CloseSend(); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	results := make([]string, 0)
	for {
		select {
		case err := <-errC:
			if err != nil {
				return echo.NewHTTPError(http.StatusBadRequest, err.Error())
			}
		case result := <-resultC:
			if result != nil {
				results = append(results, result.Result)
			}
		case <-doneC:
			return c.JSON(http.StatusOK, map[string]interface{}{
				"results": results,
			})
		}
	}
}
