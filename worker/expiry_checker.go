package worker

import (
	"fmt"
	"log"
	"time"

	"gorm.io/gorm"
	"mengonten-api/models"
)

const freePlanID = "85052fb1-7a58-4951-957f-36ce2e5588f6"

func RunExpiryCheck(db *gorm.DB, es *EmailSender) {
	log.Println("[ExpiryCheck] ===================== START =====================")
	log.Println("[ExpiryCheck] Starting expiry check...")

	expiredCount := expireOrders(db)
	if expiredCount > 0 {
		log.Printf("[ExpiryCheck] Expired %d orders", expiredCount)
		switchedCount := switchToFreePlan(db)
		log.Printf("[ExpiryCheck] Switched %d users to free plan", switchedCount)
	} else {
		log.Println("[ExpiryCheck] No orders to expire")
	}

	reminderCount := sendReminders(db, es)
	log.Printf("[ExpiryCheck] Sent %d reminder emails", reminderCount)

	log.Println("[ExpiryCheck] =====================  END  =====================")
}

func expireOrders(db *gorm.DB) int64 {
	result := db.Model(&models.Order{}).
		Joins("JOIN subscription_plans ON subscription_plans.id = orders.plan_id").
		Where("orders.status = 'active' AND orders.expired_at < NOW() AND subscription_plans.duration_days > 0").
		Update("orders.status", "expired")
	return result.RowsAffected
}

func switchToFreePlan(db *gorm.DB) int64 {
	var expiredOrders []models.Order
	db.Where("status = 'expired' AND plan_id != ?", freePlanID).Find(&expiredOrders)

	count := int64(0)
	for _, o := range expiredOrders {
		var existing models.Order
		err := db.Where("user_id = ? AND status = 'active' AND plan_id = ?",
			o.UserID, freePlanID).First(&existing).Error
		if err != nil {
			var freePlan models.SubscriptionPlan
			if db.First(&freePlan, freePlanID).Error != nil {
				continue
			}

			now := time.Now()

			dummyTxn := models.Transaction{
				UserID:             o.UserID,
				SubscriptionPlanID: freePlan.ID,
				ReferenceID:        fmt.Sprintf("FREE-%d-%s", now.UnixNano(), o.UserID.String()[:8]),
				ProductName:        freePlan.Name,
				PaymentTotal:       0,
				Status:             "success",
				PaymentAt:          &now,
			}
			if db.Create(&dummyTxn).Error != nil {
				continue
			}

			newOrder := models.Order{
				UserID:        o.UserID,
				TransactionID: dummyTxn.ID,
				PlanID:        freePlan.ID,
				ProductName:   freePlan.Name,
				ProductPrice:  freePlan.FinalPrice,
				Status:        "active",
				ExpiredAt:     now.AddDate(0, 0, freePlan.DurationDays),
			}
			if db.Create(&newOrder).Error != nil {
				db.Delete(&dummyTxn)
				continue
			}

			db.Model(&dummyTxn).Update("order_id", newOrder.ID)

			var rules []models.SubscriptionRule
			db.Where("plan_id = ?", freePlan.ID).Find(&rules)
			for _, r := range rules {
				ruleID := r.ID
				orderRule := models.OrderRule{
					OrderID:            newOrder.ID,
					SubscriptionRuleID: &ruleID,
					RuleKey:            r.RuleKey,
					RuleValue:          r.RuleValue,
				}
				db.Create(&orderRule)
			}

			count++
		}
	}
	return count
}

func sendReminders(db *gorm.DB, es *EmailSender) int {
	if es == nil {
		return 0
	}

	var orders []models.Order
	db.Model(&models.Order{}).
		Joins("JOIN subscription_plans ON subscription_plans.id = orders.plan_id").
		Where("orders.status = 'active' AND subscription_plans.duration_days > 0").
		Where("orders.expired_at BETWEEN NOW() AND NOW() + INTERVAL '7 days'").
		Find(&orders)

	count := 0
	for _, o := range orders {
		daysLeft := int(time.Until(o.ExpiredAt).Hours() / 24)
		if daysLeft < 0 {
			continue
		}

		shouldSend := false
		if daysLeft == 7 || daysLeft == 3 || daysLeft == 1 {
			if o.LastReminderSentAt == nil {
				shouldSend = true
			} else {
				lastSent := time.Since(*o.LastReminderSentAt)
				if (daysLeft == 7 && lastSent.Hours() > 12) ||
					(daysLeft == 3 && lastSent.Hours() > 12) ||
					(daysLeft == 1 && lastSent.Hours() > 12) {
					shouldSend = true
				}
			}
		}

		if !shouldSend {
			continue
		}

		var user models.User
		if db.First(&user, o.UserID).Error != nil {
			continue
		}

		expiredAt := o.ExpiredAt.Format("02 Jan 2006, 15:04 WIB")
		var subject string
		switch daysLeft {
		case 7:
			subject = "Langganan Anda akan berakhir dalam 7 hari"
		case 3:
			subject = "Langganan Anda akan berakhir dalam 3 hari"
		case 1:
			subject = "Langganan Anda berakhir besok! Perpanjang sekarang"
		default:
			subject = fmt.Sprintf("Langganan akan berakhir dalam %d hari", daysLeft)
		}

		err := es.SendExpiryReminder(user.Email, o.PlanID.String(), o.ProductName, expiredAt, daysLeft, subject, o.ProductPrice)
		if err != nil {
			log.Printf("[ExpiryCheck] Failed to send reminder to %s: %v", user.Email, err)
			continue
		}

		now := time.Now()
		db.Model(&o).Update("last_reminder_sent_at", &now)
		count++
	}
	return count
}
