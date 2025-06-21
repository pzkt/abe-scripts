package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/pzkt/abe-scripts/abe-scheme/internal/utils"
	pb "github.com/pzkt/abe-scripts/abe-scheme/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

/*
docker setup commands:

docker run --name postgres-container -e POSTGRES_PASSWORD=pwd -p 5432:5432 -d postgres

docker start -a postgres-container

docker run --name pgadmin -p 15432:80 -e 'PGADMIN_DEFAULT_EMAIL=user@domain.com' -e 'PGADMIN_DEFAULT_PASSWORD=pwd' -d dpage/pgadmin4

Host name/address: 172.17.0.2
Port: 5432
Maintenance database: postgres
Username: postgres
Password: pwd

*/

type DatabaseService struct {
	db *sql.DB
}

type RecordServiceServer struct {
	dbService *DatabaseService
	// forward compatibility
	pb.UnimplementedRecordServiceServer
}

func main() {
	dbPassword := "pwd"
	connection := fmt.Sprintf("postgres://postgres:%s@localhost:5432/data?sslmode=disable", dbPassword)

	db := utils.Assure(sql.Open("postgres", connection))
	utils.Try(db.Ping())

	defer db.Close()

	dbService := &DatabaseService{db: db}

	lis := utils.Assure(net.Listen("tcp", ":50051"))

	grpcServer := grpc.NewServer()
	pb.RegisterRecordServiceServer(grpcServer, &RecordServiceServer{dbService: dbService})
	log.Println("Database running on port :50051")
	grpcServer.Serve(lis)

	setup(db)
}

func setup(db *sql.DB) {
	//create the key-value table for table row relations
	query := `CREATE TABLE IF NOT EXISTS relations (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		private_write_key BYTEA,
		public_write_key BYTEA,
		data BYTEA,
		created TIMESTAMP DEFAULT NOW()
	)`

	utils.Assure(db.Exec(query))

	query = `CREATE TABLE IF NOT EXISTS table_one (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		private_write_key BYTEA,
		public_write_key BYTEA,
		data BYTEA,
		created TIMESTAMP DEFAULT NOW()
	)`

	utils.Assure(db.Exec(query))

	query = `CREATE TABLE IF NOT EXISTS table_two (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		private_write_key BYTEA,
		public_write_key BYTEA,
		data BYTEA,
		created TIMESTAMP DEFAULT NOW()
	)`

	utils.Assure(db.Exec(query))
}

func (s *RecordServiceServer) AddEntry(ctx context.Context, req *pb.AddEntryRequest) (*pb.AddEntryResponse, error) {

	query := fmt.Sprintf(
		`INSERT INTO %s (private_write_key, public_write_key, data) 
         VALUES ($1, $2, $3) 
         RETURNING id, created`,
		req.Table,
	)

	var id uuid.UUID
	var created time.Time

	err := s.dbService.db.QueryRowContext(ctx, query,
		req.WriteKeyCipher,
		req.MarshaledPublicWriteKey,
		req.DataCipher,
	).Scan(&id, &created)

	if err != nil {
		log.Printf("Insert failed: %v", err)
		return nil, status.Error(codes.Internal, "failed to insert record")
	}

	return &pb.AddEntryResponse{
		Id:      id.String(),
		Created: timestamppb.New(created),
	}, nil
}

func (s *RecordServiceServer) GetEntry(ctx context.Context, req *pb.GetEntryRequest) (*pb.GetEntryResponse, error) {
	// Convert string UUID to PostgreSQL UUID type
	rowID, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid UUID format")
	}

	var data []byte
	err = s.dbService.db.QueryRowContext(ctx,
		`SELECT data FROM table_one WHERE id = $1`,
		rowID,
	).Scan(&data)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Errorf(codes.NotFound, "record not found")
		}
		return nil, status.Errorf(codes.Internal, "database error: %v", err)
	}

	return &pb.GetEntryResponse{Data: data}, nil
}
