import Link from 'next/link';

type InfoAction = {
  label: string;
  href: string;
  primary?: boolean;
};

type StaticInfoPageProps = {
  title: string;
  description: string;
  badge: string;
  sections: Array<{
    title: string;
    items: string[];
  }>;
  actions: InfoAction[];
};

export default function StaticInfoPage({
  title,
  description,
  badge,
  sections,
  actions,
}: StaticInfoPageProps) {
  return (
    <main className='min-h-screen bg-[radial-gradient(circle_at_top,rgba(229,9,20,0.18),transparent_35%),#090909] px-4 py-10 text-white sm:px-8'>
      <div className='mx-auto max-w-4xl'>
        <div className='mb-8 rounded-3xl border border-white/10 bg-white/[0.03] p-6 backdrop-blur sm:p-10'>
          <div className='mb-4 inline-flex rounded-full bg-red-600/15 px-4 py-1 text-sm font-medium text-red-300'>
            {badge}
          </div>
          <h1 className='text-3xl font-black sm:text-5xl'>{title}</h1>
          <p className='mt-4 max-w-2xl text-sm leading-7 text-zinc-300 sm:text-base'>
            {description}
          </p>

          <div className='mt-6 flex flex-wrap gap-3'>
            {actions.map((action) => (
              <Link
                key={action.href}
                href={action.href}
                className={`rounded-full px-5 py-3 text-sm font-semibold transition-colors ${
                  action.primary
                    ? 'bg-white text-black hover:bg-zinc-200'
                    : 'border border-white/15 bg-white/5 text-white hover:bg-white/10'
                }`}
              >
                {action.label}
              </Link>
            ))}
          </div>
        </div>

        <div className='grid gap-4 md:grid-cols-2'>
          {sections.map((section) => (
            <section
              key={section.title}
              className='rounded-2xl border border-white/10 bg-zinc-950/80 p-5'
            >
              <h2 className='text-lg font-bold text-white'>{section.title}</h2>
              <ul className='mt-4 space-y-3 text-sm text-zinc-300'>
                {section.items.map((item) => (
                  <li key={item} className='flex gap-2 leading-6'>
                    <span className='mt-1 text-red-400'>•</span>
                    <span>{item}</span>
                  </li>
                ))}
              </ul>
            </section>
          ))}
        </div>
      </div>
    </main>
  );
}
