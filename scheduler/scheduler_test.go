package scheduler

import (
	pb "version.uuzu.com/zhuhuipeng/djob/message"
	"reflect"
	"testing"
)

func Test_newEntery(t *testing.T) {
	type args struct {
		job *pb.Job
	}
	tests := []struct {
		name    string
		args    args
		want    *entry
		wantErr bool
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := newEntery(tt.args.job)
			if (err != nil) != tt.wantErr {
				t.Errorf("newEntery() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("newEntery() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestScheduler_deleteEntryByName(t *testing.T) {
	type fields struct {
		addEntry    chan *entry
		deleteEntry chan string
		RunJobCh    chan *pb.Job
		stopCh      chan struct{}
		running     bool
		entries     []*entry
		nameToIndex map[string]int
	}
	type args struct {
		name string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "test_delete_entry_01",
			fields: fields{},
			args: args{name:"test_delete_01"},
			wantErr: true,
		},
		{
			name:"test_delete_entry_02",
			fields: fields{
				entries:[]*entry{
					&entry{
						Job: &pb.Job{
							Name: "job1",
						},
					},
				},
				nameToIndex: map[string]int{
					"job1": 0,
				},
			},
			args: args{name:"job1"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Scheduler{
				addEntry:    tt.fields.addEntry,
				deleteEntry: tt.fields.deleteEntry,
				RunJobCh:    tt.fields.RunJobCh,
				stopCh:      tt.fields.stopCh,
				running:     tt.fields.running,
				entries:     tt.fields.entries,
				nameToIndex: tt.fields.nameToIndex,
			}
			if err := s.deleteEntryByName(tt.args.name); (err != nil) != tt.wantErr {
				t.Errorf("Scheduler.deleteEntryByName() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestScheduler_AddJob(t *testing.T) {
	type fields struct {
		addEntry    chan *entry
		deleteEntry chan string
		RunJobCh    chan *pb.Job
		stopCh      chan struct{}
		running     bool
		entries     []*entry
		nameToIndex map[string]int
	}
	type args struct {
		job *pb.Job
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Scheduler{
				addEntry:    tt.fields.addEntry,
				deleteEntry: tt.fields.deleteEntry,
				RunJobCh:    tt.fields.RunJobCh,
				stopCh:      tt.fields.stopCh,
				running:     tt.fields.running,
				entries:     tt.fields.entries,
				nameToIndex: tt.fields.nameToIndex,
			}
			if err := s.AddJob(tt.args.job); (err != nil) != tt.wantErr {
				t.Errorf("Scheduler.AddJob() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestScheduler_DeleteJob(t *testing.T) {
	type fields struct {
		addEntry    chan *entry
		deleteEntry chan string
		RunJobCh    chan *pb.Job
		stopCh      chan struct{}
		running     bool
		entries     []*entry
		nameToIndex map[string]int
	}
	type args struct {
		name string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Scheduler{
				addEntry:    tt.fields.addEntry,
				deleteEntry: tt.fields.deleteEntry,
				RunJobCh:    tt.fields.RunJobCh,
				stopCh:      tt.fields.stopCh,
				running:     tt.fields.running,
				entries:     tt.fields.entries,
				nameToIndex: tt.fields.nameToIndex,
			}
			s.DeleteJob(tt.args.name)
		})
	}
}
