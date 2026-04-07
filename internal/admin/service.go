// Package Admin contains all the handlers and services related to admin management.
package admin

import (
	"context"

	"go-server/internal/database"
)

type Service interface {
	GetKpis(ctx context.Context) error
	GetMonthlyStats(ctx context.Context) error
	GetCategoryDistribution(ctx context.Context) error
	GetClusterStats(ctx context.Context) error
}

type service struct {
	db database.Service
}

func NewService(db database.Service) Service {
	return &service{
		db: db,
	}
}

func (s *service) GetKpis(ctx context.Context) error {
	// getKpis method from the old server codebase:
	// const [totalTrainings, totalParticipants, completedCount, enrolledCount] =
	//      await Promise.all([
	//        this.prisma.training.count({ where: { isActive: true } }),
	//        this.prisma.nomination.count({
	//          where: { status: { in: ACTIVE_STATUSES } },
	//        }),
	//        this.prisma.nomination.count({
	//          where: { status: NominationStatus.COMPLETED },
	//        }),
	//        this.prisma.nomination.count({
	//          where: { status: { in: ACTIVE_STATUSES } },
	//        }),
	//      ]);
	//
	//    const avgAttendance =
	//      enrolledCount > 0
	//        ? Math.round((completedCount / enrolledCount) * 100)
	//        : 0;
	//
	//    // Man-days: sum of (endDate - startDate + 1) * active nominations
	//    const trainings = await this.prisma.training.findMany({
	//      where: { isActive: true },
	//      include: {
	//        _count: {
	//          select: {
	//            nominations: {
	//              where: { status: { in: ACTIVE_STATUSES } },
	//            },
	//          },
	//        },
	//      },
	//    });
	//
	//    const totalManDays = trainings.reduce((sum, t) => {
	//      const durationDays =
	//        Math.ceil((t.endDate.getTime() - t.startDate.getTime()) / 86400000) + 1;
	//      return sum + durationDays * t._count.nominations;
	//    }, 0);
	//
	//    return {
	//      totalTrainings,
	//      totalParticipants,
	//      avgAttendance,
	//      totalManDays,
	//    };
}

func (s *service) GetMonthlyStats(ctx context.Context) error {
	// GetMonthlyStas method from the old server codebase:
	// const nominations = await this.prisma.nomination.findMany({
	//   where: {
	//     status: { in: ACTIVE_STATUSES },
	//     trainingId: { not: null },
	//   },
	//   include: {
	//     training: { select: { startDate: true } },
	//   },
	// });
	//
	// const monthMap: Record<
	//   string,
	//   { month: string; participants: number; trainings: Set<string> }
	// > = {};
	//
	// for (const nom of nominations) {
	//   if (!nom.training) continue;
	//   const date = nom.training.startDate;
	//   const key = `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, '0')}`;
	//   const label = date.toLocaleString('default', {
	//     month: 'short',
	//     year: '2-digit',
	//   });
	//   if (!monthMap[key]) {
	//     monthMap[key] = { month: label, participants: 0, trainings: new Set() };
	//   }
	//   monthMap[key].participants++;
	//   monthMap[key].trainings.add(nom.trainingId!);
	// }
	//
	// return Object.entries(monthMap)
	//   .sort(([a], [b]) => a.localeCompare(b))
	//   .map(([, v]) => ({
	//     month: v.month,
	//     participants: v.participants,
	//     trainings: v.trainings.size,
	//   }));
}

func (s *service) GetCategoryDistribution(ctx context.Context) error {
	// getCategoryDistribution method from the old server codebase:
	// const result = await this.prisma.nomination.groupBy({
	//      by: ['trainingId'],
	//      where: {
	//        status: { in: ACTIVE_STATUSES },
	//        trainingId: { not: null },
	//      },
	//      _count: true,
	//    });
	//
	//    const trainingIds = result
	//      .map((r) => r.trainingId)
	//      .filter((id): id is string => id !== null);
	//    const trainings = await this.prisma.training.findMany({
	//      where: { id: { in: trainingIds } },
	//      select: { id: true, category: true },
	//    });
	//
	//    const categoryMap: Record<string, number> = {};
	//    for (const r of result) {
	//      const training = trainings.find((t) => t.id === r.trainingId);
	//      if (!training) continue;
	//      categoryMap[training.category] =
	//        (categoryMap[training.category] ?? 0) + r._count;
	//    }
	//
	//    return Object.entries(categoryMap).map(([name, value]) => ({
	//      name,
	//      value,
	//    }));
}

func (s *service) GetClusterStats(ctx context.Context) error {
	// getClusterStats method from the old server codebase:
	// const users = await this.prisma.user.findMany({
	//      where: { isActive: true, role: { not: 'ADMIN' } },
	//      select: {
	//        id: true,
	//        cluster: true,
	//        nominations: {
	//          where: { status: { in: ACTIVE_STATUSES } },
	//          select: { id: true },
	//        },
	//      },
	//    });
	//
	//    const clusterMap: Record<string, { total: number; trained: Set<string> }> =
	//      {};
	//    for (const user of users) {
	//      const cluster = user.cluster ?? 'Unassigned';
	//      if (!clusterMap[cluster]) {
	//        clusterMap[cluster] = { total: 0, trained: new Set() };
	//      }
	//      clusterMap[cluster].total++;
	//      if (user.nominations.length > 0) {
	//        clusterMap[cluster].trained.add(user.id);
	//      }
	//    }
	//
	//    return Object.entries(clusterMap).map(([cluster, data]) => ({
	//      cluster,
	//      totalEmployees: data.total,
	//      trained: data.trained.size,
	//      untrained: data.total - data.trained.size,
	//    }));
	//  }
}
