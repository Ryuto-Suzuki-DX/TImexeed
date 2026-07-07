"use client";

type PageTitleProps = {
  title: string;
  description?: string;
  className?: string;
};

export default function PageTitle({
  title,
  description,
  className,
}: PageTitleProps) {
  const wrapperClassName = [
    "timexeed-page-title",
    className ?? "",
  ]
    .filter(Boolean)
    .join(" ");

  return (
    <>
      <div className={wrapperClassName}>
        <h1 className="timexeed-page-title-heading">{title}</h1>

        {description && (
          <p className="timexeed-page-title-description">{description}</p>
        )}
      </div>

      <style>{`
        .timexeed-page-title {
          margin-bottom: 24px;
        }

        .timexeed-page-title-heading {
          margin: 0 0 8px;
          color: #ea580c;
          font-size: 32px;
          font-weight: bold;
          line-height: 1.25;
          overflow-wrap: anywhere;
        }

        .timexeed-page-title-description {
          margin: 0;
          color: #666666;
          font-size: 14px;
          line-height: 1.6;
          overflow-wrap: anywhere;
        }

        @media (max-width: 768px) {
          .timexeed-page-title {
            margin-bottom: 10px;
          }

          .timexeed-page-title-heading {
            margin-bottom: 4px;
            font-size: 20px;
            line-height: 1.2;
          }

          .timexeed-page-title-description {
            font-size: 10px;
            line-height: 1.4;
          }
        }

        @media (max-width: 480px) {
          .timexeed-page-title {
            margin-bottom: 8px;
          }

          .timexeed-page-title-heading {
            margin-bottom: 3px;
            font-size: 18px;
          }

          .timexeed-page-title-description {
            font-size: 9px;
            line-height: 1.35;
          }
        }
      `}</style>
    </>
  );
}
