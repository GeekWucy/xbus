package configs

import (
	"database/sql"
	"github.com/gocomm/dbutil"
	"github.com/golang/glog"
	"github.com/infrmods/xbus/utils"
	"time"
)

const (
	ConfigStatusOk      = 0
	ConfigStatusDeleted = -1
)

type DBConfigItem struct {
	Id         int64     `json:"id"`
	Status     int       `json:"-"`
	Name       string    `json:"name"`
	Value      string    `json:"value"`
	CreateTime time.Time `json:"create_time"`
	ModifyTime time.Time `json:"modify_time"`
}

func GetDBConfigCount(db *sql.DB, prefix string) (int64, error) {
	var count int64
	var err error
	q := `select count(*) from configs`
	if prefix == "" {
		err = dbutil.Query(db, &count, q)
	} else {
		err = dbutil.Query(db, &count, q+` where name like ?`, prefix+"%")
	}

	return count, err
}

func ListDBConfigs(db *sql.DB, prefix string, skip, limit int) ([]string, error) {
	args := make([]interface{}, 0, 3)
	q := `select name from configs where status=?`
	args = append(args, ConfigStatusOk)
	if prefix != "" {
		q += ` and name like ?`
		args = append(args, prefix+"%")
	}
	q += ` order by modify_time desc limit ?,?`
	args = append(args, skip)
	args = append(args, limit)

	var items []string
	if err := dbutil.Query(db, &items, q, args...); err == nil {
		return items, nil
	} else {
		return nil, err
	}
}

type ConfigHistory struct {
	Id         int64     `json:"id"`
	Name       string    `json:"name"`
	AppId      int64     `json:"modified_by"`
	Value      string    `json:"value"`
	CreateTime time.Time `json:"create_time"`
}

func (ctrl *ConfigCtrl) setDBConfig(name string, appId int64, value string) (rerr error) {
	tx, err := ctrl.db.Begin()
	if err != nil {
		glog.Errorf("new db tx fail: %v", err)
		return utils.NewError(utils.EcodeSystemError, "new db tx fail")
	}

	defer func() {
		if rerr != nil {
			if err := tx.Rollback(); err != nil {
				glog.Warningf("tx roolback fail: %v", err)
			}
		}
	}()

	if _, err := tx.Exec(`insert into configs(status,name,value,create_time,modify_time)
                          values(?,?,?,now(),now())
                          on duplicate key update status=?, value=?, modify_time=now()`,
		ConfigStatusOk, name, value, ConfigStatusOk, value); err != nil {
		glog.Errorf("insert db config(%s) fail: %v", name, err)
		return utils.NewError(utils.EcodeSystemError, "update db config fail")
	}
	if _, err := tx.Exec(`insert into config_histories(name,app_id,value,create_time)
                          values(?,?,?,now())`, name, appId, value); err != nil {
		glog.Errorf("insert db config history fail: %v", err)
		return utils.NewError(utils.EcodeSystemError, "insert db config history fail")
	}

	if err := tx.Commit(); err == nil {
		return nil
	} else {
		glog.Errorf("set db config(%s), commit fail: %v", name, err)
		return utils.NewError(utils.EcodeSystemError, "commit db fail")
	}
}

func (ctrl *ConfigCtrl) deleteDBConfig(name string) error {
	if _, err := ctrl.db.Exec(`update configs set status=? where name=?`, ConfigStatusDeleted, name); err != nil {
		return utils.NewError(utils.EcodeSystemError, "delete config fail")
	}
	return nil
}

type AppConfigState struct {
	Id         int64     `json:"id"`
	AppId      int64     `json:"app_id"`
	AppNode    string    `json:"app_node"`
	ConfigName string    `json:"config_name"`
	Version    int64     `json:"version"`
	CreateTime time.Time `json:"create_time"`
	ModifyTime time.Time `json:"modify_time"`
}

func (ctrl *ConfigCtrl) changeAppConfigState(appId int64, appNode, configName string, version int64) error {
	if appId <= 0 {
		return nil
	}

	if _, err := ctrl.db.Exec(`insert into app_config_states(app_id,app_node,config_name,version,create_time,modify_time)
                            values(?,?,?,?,now(),now())
                            on duplicate key update version=?`,
		appId, appNode, configName, version, version); err != nil {
		glog.Errorf("change app(%d - %s) config(%s) state(ver: %d) fail: %v", appId, appNode, configName, version, err)
		return utils.NewError(utils.EcodeSystemError, "change app config state fail")
	} else {
		return nil
	}
}
