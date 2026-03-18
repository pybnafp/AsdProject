/**
 * 操作日志记录管理-服务类
 * @author Evotrek 研发团队
 * @since 2024-11-10
 * @File : oper_log
 */
package services

import (
	"asd/app/dto"
	"asd/app/models"
	"asd/app/vo"
	"asd/utils/gconv"
	"errors"
	"strings"
	"time"

	"github.com/beego/beego/v2/client/orm"
)

// 中间件管理服务
var OperLog = new(operLogService)

type operLogService struct{}

func (s *operLogService) GetList(req dto.OperLogPageReq) ([]vo.OperLogInfoVo, int64, error) {
	// 初始化查询实例
	query := orm.NewOrm().QueryTable(new(models.OperLog)).Filter("mark", 1)

	// 操作类型：0其它 1新增 2修改 3删除 4查询 5设置状态 6导入 7导出 8设置权限 9设置密码

	if req.OperType > 0 {
		query = query.Filter("oper_type", req.OperType)
	}

	// 日志状态：0正常日志 1错误日志

	if req.Status > 0 {
		query = query.Filter("status", req.Status)
	}

	// 排序
	query = query.OrderBy("-id")
	// 查询总数
	count, _ := query.Count()
	// 分页设置
	offset := (req.Page - 1) * req.Limit
	query = query.Limit(req.Limit, offset)
	// 查询列表
	lists := make([]models.OperLog, 0)
	// 对象转换
	query.All(&lists)

	// 数据处理
	var result []vo.OperLogInfoVo
	for _, v := range lists {
		item := vo.OperLogInfoVo{}
		item.OperLog = v

		result = append(result, item)
	}

	// 返回结果
	return result, count, nil
}

func (s *operLogService) Add(req dto.OperLogAddReq, userId int) (int64, error) {
	// 实例化对象
	var entity models.OperLog

	entity.Model = req.Model

	entity.OperType = req.OperType

	entity.UserId = req.UserId
	entity.OperMethod = req.OperMethod
	entity.Username = req.Username
	entity.OperName = req.OperName
	entity.OperUrl = req.OperUrl
	entity.OperIp = req.OperIp
	entity.OperLocation = req.OperLocation
	entity.RequestParam = req.RequestParam
	entity.Result = req.Result

	entity.Status = int8(req.Status)

	entity.UserAgent = req.UserAgent
	entity.Note = req.Note
	entity.CreateUser = userId
	entity.CreateTime = time.Now()
	entity.UpdateUser = userId
	entity.UpdateTime = time.Now()
	entity.Mark = 1
	// 插入数据
	return entity.Insert()
}

func (s *operLogService) Update(req dto.OperLogUpdateReq, userId int) (int64, error) {
	// 查询记录
	entity := &models.OperLog{Id: req.Id}
	err := entity.Get()
	if err != nil {
		return 0, errors.New("记录不存在")
	}

	entity.Model = req.Model

	entity.OperType = req.OperType

	entity.OperMethod = req.OperMethod
	entity.Username = req.Username
	entity.OperName = req.OperName
	entity.OperUrl = req.OperUrl
	entity.OperIp = req.OperIp
	entity.OperLocation = req.OperLocation
	entity.RequestParam = req.RequestParam
	entity.Result = req.Result

	entity.Status = int8(req.Status)

	entity.UserAgent = req.UserAgent
	entity.Note = req.Note
	entity.UpdateUser = userId
	entity.UpdateTime = time.Now()
	// 更新记录
	return entity.Update()
}

// 删除
func (s *operLogService) Delete(ids string) (int64, error) {
	// 记录ID
	idsArr := strings.Split(ids, ",")
	if len(idsArr) == 1 {
		// 单个删除
		entity := &models.OperLog{Id: gconv.Int(ids)}
		rows, err := entity.Delete()
		if err != nil || rows == 0 {
			return 0, errors.New("删除失败:" + err.Error())
		}
		return rows, nil
	} else {
		// 批量删除
		count := 0
		for _, v := range idsArr {
			entity := &models.OperLog{Id: gconv.Int(v)}
			rows, err := entity.Delete()
			if err != nil || rows == 0 {
				continue
			}
			count++
		}
		return int64(count), nil
	}
}

func (s *operLogService) Status(req dto.OperLogStatusReq, userId int) (int64, error) {
	// 查询记录是否存在
	entity := &models.OperLog{Id: req.Id}
	err := entity.Get()
	if err != nil {
		return 0, errors.New("记录不存在")
	}
	entity.Status = int8(req.Status)
	entity.UpdateUser = userId
	entity.UpdateTime = time.Now()
	return entity.Update()
}
